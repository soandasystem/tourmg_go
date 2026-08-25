package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const signatureRelID = "rIdSig1"
const signatureMediaPath = "word/media/signature.png"
const signatureRelTarget = "media/signature.png"

// drawingXMLTemplate es la plantilla OpenXML para insertar la imagen PNG en un archivo Word (.docx)
var drawingXMLTemplate = fmt.Sprintf(`<w:drawing xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">
  <wp:inline distT="0" distB="0" distL="0" distR="0">
    <wp:extent cx="1800000" cy="900000"/>
    <wp:effectExtent l="0" t="0" r="0" b="0"/>
    <wp:docPr id="999" name="Firma"/>
    <wp:cNvGraphicFramePr>
      <a:graphicFrameLocks noChangeAspect="1"/>
    </wp:cNvGraphicFramePr>
    <a:graphic>
      <a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">
        <pic:pic>
          <pic:nvPicPr>
            <pic:cNvPr id="0" name="firma.png"/>
            <pic:cNvPicPr/>
          </pic:nvPicPr>
          <pic:blipFill>
            <a:blip r:embed="%s"/>
            <a:stretch>
              <a:fillRect/>
            </a:stretch>
          </pic:blipFill>
          <pic:spPr>
            <a:xfrm>
              <a:off x="0" y="0"/>
              <a:ext cx="1800000" cy="900000"/>
            </a:xfrm>
            <a:prstGeom prst="rect">
              <a:avLst/>
            </a:prstGeom>
          </pic:spPr>
        </pic:pic>
      </a:graphicData>
    </a:graphic>
  </wp:inline>
</w:drawing>`, signatureRelID)

// InjectSignatureImage reemplaza marcas como {{firma}}, {firma}, [FIRMA] en un DOCX
// por la imagen PNG de la firma directamente en el XML de OpenXML.
func InjectSignatureImage(docxBytes []byte, firmaPngBytes []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return nil, fmt.Errorf("error leyendo ZIP del DOCX: %w", err)
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	foundPlaceholder := false

	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("error abriendo archivo %s dentro del DOCX: %w", file.Name, err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("error leyendo contenido de %s: %w", file.Name, err)
		}

		switch file.Name {
		case "word/document.xml":
			docXML := string(content)

			// 1. Intentar reemplazos directos de cadenas
			patterns := []string{
				"{{firma}}", "{{Firma}}", "{{FIRMA}}",
				"{firma}", "{Firma}", "{FIRMA}",
				"[FIRMA]", "[firma]", "[Firma]",
			}

			for _, p := range patterns {
				if strings.Contains(docXML, p) {
					docXML = strings.ReplaceAll(docXML, p, drawingXMLTemplate)
					foundPlaceholder = true
				}
			}

			// 2. Si no se encontró por coincidencia exacta (posibles fragmentos de runs XML en Word)
			if !foundPlaceholder {
				// Buscar patrón de etiqueta fragmentada entre etiquetas <w:t>...</w:t>
				rx := regexp.MustCompile(`(?i)<w:t[^>]*>[\s\S]*?\{{1,2}\s*firma\s*\}}{1,2}[\s\S]*?</w:t>`)
				if rx.MatchString(docXML) {
					docXML = rx.ReplaceAllString(docXML, "<w:t></w:t>"+drawingXMLTemplate)
					foundPlaceholder = true
				}
			}

			// Escribir document.xml modificado
			w, err := zipWriter.Create(file.Name)
			if err != nil {
				return nil, err
			}
			if _, err := w.Write([]byte(docXML)); err != nil {
				return nil, err
			}

		case "word/_rels/document.xml.rels":
			relsXML := string(content)
			relEntry := fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>`, signatureRelID, signatureRelTarget)

			if !strings.Contains(relsXML, fmt.Sprintf(`Id="%s"`, signatureRelID)) {
				relsXML = strings.Replace(relsXML, "</Relationships>", relEntry+"</Relationships>", 1)
			}

			w, err := zipWriter.Create(file.Name)
			if err != nil {
				return nil, err
			}
			if _, err := w.Write([]byte(relsXML)); err != nil {
				return nil, err
			}

		case "[Content_Types].xml":
			typesXML := string(content)
			pngTypeEntry := `<Default Extension="png" ContentType="image/png"/>`

			if !strings.Contains(typesXML, `Extension="png"`) {
				typesXML = strings.Replace(typesXML, "</Types>", pngTypeEntry+"</Types>", 1)
			}

			w, err := zipWriter.Create(file.Name)
			if err != nil {
				return nil, err
			}
			if _, err := w.Write([]byte(typesXML)); err != nil {
				return nil, err
			}

		default:
			// Copiar el resto de archivos sin modificación
			w, err := zipWriter.Create(file.Name)
			if err != nil {
				return nil, err
			}
			if _, err := w.Write(content); err != nil {
				return nil, err
			}
		}
	}

	if !foundPlaceholder {
		return nil, fmt.Errorf("no se encontró el marcador {{firma}} en la plantilla DOCX")
	}

	// Inyectar el archivo de imagen PNG en word/media/signature.png
	imgWriter, err := zipWriter.Create(signatureMediaPath)
	if err != nil {
		return nil, fmt.Errorf("error creando archivo de imagen en ZIP: %w", err)
	}
	if _, err := imgWriter.Write(firmaPngBytes); err != nil {
		return nil, fmt.Errorf("error escribiendo imagen en ZIP: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("error cerrando archivo ZIP: %w", err)
	}

	return buf.Bytes(), nil
}
