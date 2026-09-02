package handlers

import (
	"context"
	"fmt"
	"net/http"
	"tourmanager/config"
	"tourmanager/core/models"
	"tourmanager/core/ports"
	"tourmanager/util"

	"github.com/gin-gonic/gin"
)

// SetAuthRoutes define las rutas de autenticación públicas
// En este caso inyectamos los servicios necesarios para validar contra la BD
func SetAuthRoutes(ctx context.Context, cfg config.Config, router *gin.Engine, userSvc ports.UsersService, cursoSvc ports.CursoService, saleSvc ports.SaleService) {
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/login", loginHandler(ctx, userSvc, cursoSvc, saleSvc))
	}
}

// loginHandler recibe la petición del frontend y según el tipo decide qué validar
func loginHandler(ctx context.Context, userSvc ports.UsersService, cursoSvc ports.CursoService, saleSvc ports.SaleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqCtx := c.Request.Context() // Este contexto contiene el schema inyectado por el middleware

		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Faltan datos o el formato es incorrecto"})
			return
		}

		var token string
		var err error

		// Dependiendo del login_type, realizamos la validación en la base de datos
		switch req.LoginType {
		case util.AuthTypeUser:
			if userSvc == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Servicio de usuarios no configurado"})
				return
			}
			// Buscar en la BD mediante el servicio
			filter := map[string]interface{}{
				"username": req.Username,
				"password": req.Password, // Nota: en producción, no almacenes ni filtres por texto plano. Debes hashear.
			}
			usersList, errSvc := userSvc.GetAll(reqCtx, filter)
			if errSvc != nil || usersList == nil || len(usersList.Items) == 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales de usuario inválidas"})
				return
			}

			// Tomamos el primer usuario encontrado
			user := usersList.Items[0]
			roleIDStr := fmt.Sprintf("%d", user.RolesId)

			token, err = util.GenerateUserToken(user.ID, roleIDStr)

		case util.AuthTypeCourse:
			if cursoSvc == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Servicio de cursos no configurado"})
				return
			}
			// Buscar en la BD
			filter := map[string]interface{}{
				"rutapod":  req.Rutapod,
				"password": req.Password,
			}
			cursosList, errSvc := cursoSvc.GetAll(reqCtx, filter)
			if errSvc != nil || cursosList == nil || len(cursosList.Items) == 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales de curso inválidas"})
				return
			}

			curso := cursosList.Items[0]
			token, err = util.GenerateCourseToken(curso.ID, curso.Rutapod)

		case util.AuthTypeAccess:
			if saleSvc == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Servicio de ventas (sales) no configurado"})
				return
			}
			// Buscar la venta por accesscode
			filter := map[string]interface{}{
				"accesscode": req.AccessCode,
			}
			salesList, errSvc := saleSvc.GetAll(reqCtx, filter)
			if errSvc != nil || salesList == nil || len(salesList.Items) == 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Código de acceso inválido"})
				return
			}

			// Tomamos la primera coincidencia
			sale := salesList.Items[0]
			token, err = util.GenerateAccessCodeToken(sale.ID)

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de login no soportado. Debe ser 'user', 'course' o 'access_code'"})
			return
		}

		// Si ocurrió un error interno al firmar el token
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar credenciales"})
			return
		}

		// Devolver el token al frontend
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"type":  req.LoginType,
		})
	}
}
