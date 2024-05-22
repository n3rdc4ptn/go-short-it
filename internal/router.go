package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type TokenClaims struct {
	User string `json:"user"`
	jwt.RegisteredClaims
}

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		authorization := c.Request.Header.Get("Authorization")
		if authorization == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		splitted := strings.Split(authorization, " ")
		if len(splitted) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Bad token"})
			return
		}

		token_raw := splitted[1]
		token, err := jwt.ParseWithClaims(token_raw, &TokenClaims{},
			func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				secret := viper.GetString("app_secret")
				return []byte(secret), nil
			})

		switch {
		case token.Valid:
		case errors.Is(err, jwt.ErrTokenSignatureInvalid) || errors.Is(err, jwt.ErrTokenMalformed):
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			return
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			// Token is either expired or not active yet
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token expired"})
			return
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(*TokenClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid tokens"})
			return
		}

		user, err := GetUser(db, claims.User)
		if err == nil {
			c.Set("user", user)

			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user"})
	})

	// Redirect to target
	r.GET("/:slug", func(c *gin.Context) {
		var link Link

		if err := c.ShouldBindUri(&link); err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}

		row := db.QueryRow("SELECT * FROM links WHERE slug = ?", link.Slug)
		if err := row.Scan(&link.Slug, &link.Target); err != nil {
			c.JSON(404, gin.H{"message": "Not found"})
			return
		}

		c.Redirect(307, link.Target)
	})

	r.GET("/", func(c *gin.Context) {
		rows, err := db.Query("SELECT * FROM links")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to execute query: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		var links []Link = make([]Link, 0)

		for rows.Next() {
			var link Link

			if err := rows.Scan(&link.Slug, &link.Target); err != nil {
				fmt.Println("Error scanning row:", err)
				return
			}

			links = append(links, link)
		}

		if err := rows.Err(); err != nil {
			fmt.Println("Error during rows iteration:", err)
		}

		c.JSON(200, links)
	})

	r.POST("/", func(c *gin.Context) {
		var link Link
		c.Bind(&link)

		db.Exec(`INSERT INTO links (slug, target) VALUES (?,?)
			ON CONFLICT(slug) DO UPDATE SET target=?
		`, link.Slug, link.Target, link.Target)

		c.JSON(200, link)
	})

	r.DELETE("/:slug", func(c *gin.Context) {
		var link Link

		if err := c.ShouldBindUri(&link); err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}

		_, err := db.Exec("DELETE FROM links WHERE slug = ?", link.Slug)
		if err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "Deleted " + link.Slug,
		})

	})

	return r
}
