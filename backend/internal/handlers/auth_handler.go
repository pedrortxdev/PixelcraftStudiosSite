package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	service       *service.AuthService
	emailService  *service.EmailService
	roleService   *service.RoleService
	jwtExpiration time.Duration
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService, emailService *service.EmailService, roleService *service.RoleService, expiration time.Duration) *AuthHandler {
	return &AuthHandler{
		service:       authService,
		emailService:  emailService,
		roleService:   roleService,
		jwtExpiration: expiration,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Register user
	user, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Fetch user roles
	if h.roleService != nil {
		roles, err := h.roleService.GetUserRoles(c.Request.Context(), user.ID)
		if err == nil {
			user.Roles = roles
			user.HighestRole = models.GetHighestRole(roles)
		}
	}

	// Generate JWT token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return success response
	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}
	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	user, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate user"})
		return
	}

	// Fetch user roles
	if h.roleService != nil {
		roles, err := h.roleService.GetUserRoles(c.Request.Context(), user.ID)
		if err == nil {
			user.Roles = roles
			user.HighestRole = models.GetHighestRole(roles)
		}
	}

	// Generate JWT token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return success response
	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}
	c.JSON(http.StatusOK, response)
}

// generateToken creates a JWT token for the user
func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	var roleNames []string
	if len(user.Roles) > 0 {
		for _, r := range user.Roles {
			roleNames = append(roleNames, string(r))
		}
	} else if user.HighestRole != nil {
		roleNames = append(roleNames, string(*user.HighestRole))
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"is_admin": user.IsAdmin,
		"roles":    roleNames,
		"exp":      time.Now().Add(h.jwtExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.service.JWTSecret()))
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Generates a new password and sends it via email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Email address"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email inválido"})
		return
	}

	// Generate temporary reset token
	resetToken, requestCode, err := h.service.GenerateResetToken(c.Request.Context(), req.Email)
	if err != nil {
		if err.Error() == "email not found" {
			// Don't reveal if email exists for security
			c.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviamos instruções e o Código de Redefinição para ele."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar solicitação"})
		return
	}

	// Build HTML email with reset link
	resetLink := fmt.Sprintf("https://pixelcraft-studio.store/reset-password?token=%s", resetToken)
	htmlBody := h.buildPasswordResetEmailHTML(resetLink, requestCode)

	// Send email
	err = h.emailService.SendEmailHTML(c.Request.Context(), req.Email, "Recuperação de Senha - Pixelcraft Studios", htmlBody, "")
	if err != nil {
		fmt.Printf("Error sending email to %s: %v\n", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviamos instruções e o Código de Redefinição para ele."})
}

// ResetPasswordConfirmRequest represents the request to set a new password
type ResetPasswordConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordConfirm processes the token and updates the password
func (h *AuthHandler) ResetPasswordConfirm(c *gin.Context) {
	var req ResetPasswordConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	err := h.service.ResetPasswordConfirm(c.Request.Context(), req.Token, req.Code, req.NewPassword)
	if err != nil {
		if err.Error() == "Token inválido ou já utilizado" || err.Error() == "Token expirado" || err.Error() == "Código de verificação incorreto" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Senha redefinida com sucesso"})
}

// buildPasswordResetEmailHTML builds the HTML email for password reset
func (h *AuthHandler) buildPasswordResetEmailHTML(resetLink string, requestCode string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sua Nova Senha - Pixelcraft Studios</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: linear-gradient(135deg, #f5f7fa 0%%, #c3cfe2 100%%); padding: 40px 20px; min-height: 100vh; }
        .email-wrapper { max-width: 600px; margin: 0 auto; background: white; border-radius: 20px; overflow: hidden; box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15); }
        .email-header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 40px 30px; text-align: center; position: relative; }
        .logo h1 { color: white; font-size: 36px; font-weight: 700; margin-bottom: 8px; }
        .logo p { color: rgba(255, 255, 255, 0.9); font-size: 14px; }
        .icon-lock { width: 80px; height: 80px; background: rgba(255, 255, 255, 0.2); border-radius: 50%%; margin: 25px auto 0; display: flex; align-items: center; justify-content: center; }
        .icon-lock svg { width: 40px; height: 40px; fill: white; }
        .email-body { padding: 50px 40px; }
        h2 { color: #333; font-size: 28px; margin-bottom: 15px; text-align: center; }
        .greeting { color: #666; font-size: 16px; line-height: 1.6; margin-bottom: 30px; text-align: center; }
        .password-container { background: linear-gradient(135deg, #f8f9ff 0%%, #f0f2ff 100%%); border: 2px dashed #667eea; border-radius: 15px; padding: 30px; margin: 30px 0; text-align: center; }
        .cta-button { display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white !important; padding: 16px 40px; border-radius: 10px; text-decoration: none; font-weight: 600; font-size: 16px; margin: 20px auto; text-align: center; }
        .copy-hint { color: #999; font-size: 13px; margin-top: 12px; font-style: italic; }
        .instructions { background: #fff9e6; border-left: 4px solid #ffd93d; padding: 20px; border-radius: 8px; margin: 30px 0; }
        .instructions h3 { color: #e6a500; font-size: 16px; margin-bottom: 12px; }
        .instructions ul { margin-left: 20px; color: #666; line-height: 1.8; }
        .security-note { background: #f8f9fa; border-radius: 10px; padding: 20px; margin-top: 30px; }
        .security-note h4 { color: #333; font-size: 15px; margin-bottom: 10px; }
        .security-note p { color: #666; font-size: 14px; line-height: 1.6; }
        .email-footer { background: #f8f9fa; padding: 30px 40px; text-align: center; border-top: 1px solid #e0e0e0; }
        .email-footer p { color: #999; font-size: 13px; line-height: 1.6; margin-bottom: 15px; }
    </style>
</head>
<body>
    <div class="email-wrapper">
        <div class="email-header">
            <div class="logo">
                <h1>Pixelcraft Studios</h1>
                <p>Criando experiências digitais</p>
            </div>
            <div class="icon-lock">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/></svg>
            </div>
        </div>
        <div class="email-body">
            <h2>Recuperação de Senha</h2>
            <p class="greeting">Olá! Recebemos sua solicitação de recuperação de senha. Utilize o código de verificação abaixo e clique no botão para criar uma senha nova.</p>
            <div class="password-container" style="border: none; background: transparent;">
                <h1 style="color: #667eea; letter-spacing: 5px; font-size: 32px;">%s</h1>
                <a href="%s" class="cta-button">Redefinir Minha Senha</a>
                <div class="copy-hint" style="margin-top:20px;">Ou copie e cole este link no seu navegador:<br/>%s</div>
            </div>
            <div class="instructions">
                <h3>⚠️ Importante</h3>
                <ul>
                    <li>Este link e código são válidos por apenas <strong>15 minutos</strong>.</li>
                    <li>Após este período, você precisará solicitar uma nova recuperação.</li>
                </ul>
            </div>
            <div class="security-note">
                <h4>🛡️ Segurança em primeiro lugar</h4>
                <p>Se você não solicitou esta alteração de senha, ignore este email. Sua conta permanece segura.</p>
            </div>
        </div>
        <div class="email-footer">
            <p><strong>Pixelcraft Studios</strong><br>São Paulo, SP<br>suporte@pixelcraft-studio.store</p>
            <p style="font-size: 12px; color: #bbb;">Você está recebendo este email porque solicitou recuperação de senha.<br>© 2024 Pixelcraft Studios. Todos os direitos reservados.</p>
        </div>
    </div>
</body>
</html>`, requestCode, resetLink, resetLink)
}