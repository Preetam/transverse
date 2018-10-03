package main

/**
 * Copyright (C) 2018 Preetam Jinka
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Preetam/siesta"
	"github.com/Preetam/transverse/metadata/client"
	"github.com/Preetam/transverse/metadata/middleware"
	"github.com/Preetam/transverse/metadata/token"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mailgun/mailgun-go.v1"
)

var (
	DevMode = false

	MetadataClient *client.ServiceClient
	templ          *template.Template
	TokenKey       string
	TokenCodec     *token.TokenCodec

	mg mailgun.Mailgun

	CDNDomain  = ""
	CDNVersion = "0.5.0"
)

func main() {
	addr := flag.String("addr", "localhost:4003", "Listen address")
	staticDir := flag.String("static-dir", "./static", "Path to static content")
	templatesDir := flag.String("templates-dir", "./templates", "Path to templates")
	metadataBaseAddr := flag.String("metadata-addr", "http://localhost:4000", "Address of metadata service")
	metadataToken := flag.String("metadata-token", "", "Token for metadata service")
	tokenKey := flag.String("token-key", "aaaaaaaaaaaaaaaa", "Key for token")
	flag.BoolVar(&DevMode, "dev-mode", DevMode, "Developer mode")

	s3Key := flag.String("s3-key", "", "S3 access key")
	s3Secret := flag.String("s3-secret", "", "S3 secret access key")
	s3Region := flag.String("s3-region", "nyc3", "S3 region")
	s3Endpoint := flag.String("s3-endpoint", "https://nyc3.digitaloceanspaces.com", "S3 endpoint")
	s3Directory := flag.String("s3-directory", "/tmp/s3", "local S3 directory")

	mgDomain := flag.String("mg-domain", "mg.transverseapp.com", "Mailgun domain")
	mgKey := flag.String("mg-key", "", "Mailgun key. Blank means emails are printed to stdout.")
	mgPublicKey := flag.String("mg-public-key", "", "Mailgun public key")

	recaptchaKey := flag.String("recaptcha-key", "", "Key for recaptcha")

	flag.Parse()

	if DevMode {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
		CDNDomain = "//d2zncawp28ewcu.cloudfront.net"
	}

	MetadataClient = client.NewServiceClient(*metadataBaseAddr, *metadataToken)
	s3Service := s3.New(session.New(aws.NewConfig().WithRegion(*s3Region).WithEndpoint(*s3Endpoint).WithCredentials(credentials.NewStaticCredentials(*s3Key, *s3Secret, ""))))
	mg = mailgun.NewMailgun(*mgDomain, *mgKey, *mgPublicKey)

	var err error
	templ, err = template.ParseGlob(filepath.Join(*templatesDir, "*"))
	if err != nil {
		log.Fatal(err)
	}

	TokenKey = *tokenKey
	TokenCodec = token.NewTokenCodec(1, *tokenKey)

	service := siesta.NewService("/")
	service.DisableTrimSlash() // required for static file handler
	service.AddPre(middleware.RequestIdentifier)

	service.Route("GET", "/", "serves index", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templ.ExecuteTemplate(w, "index", map[string]string{
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
	})

	service.Route("GET", "/terms", "serves index", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templ.ExecuteTemplate(w, "terms", map[string]string{
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
	})

	service.Route("GET", "/privacy-policy", "serves index", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templ.ExecuteTemplate(w, "privacy-policy", map[string]string{
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
	})

	service.Route("GET", "/login", "serves login page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var params siesta.Params
		password := params.Bool("password", false, "Use password based login")
		params.Parse(r.Form)
		if *password {
			templ.ExecuteTemplate(w, "login_password", map[string]string{
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
		} else {
			templ.ExecuteTemplate(w, "login_new", map[string]string{
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
		}
	})

	service.Route("GET", "/register", "serves register page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templ.ExecuteTemplate(w, "register", map[string]string{
			//"Error":      "Registration is disabled for now.",
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
	})

	service.Route("GET", "/verify", "serves verify page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var params siesta.Params
		action := params.String("action", "", "Action to verify")
		params.Parse(r.Form)
		templ.ExecuteTemplate(w, "verify", map[string]string{
			"action":     *action,
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
	})

	service.Route("POST", "/verify", "serves verify page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var params siesta.Params
		action := params.String("action", "", "Action to verify")
		tokenCode := params.Int("token", 0, "Token code")
		r.Form.Set("token", strings.TrimLeft(r.Form.Get("token"), "0"))
		err := params.Parse(r.Form)
		if err != nil {
			log.Println(err)
			templ.ExecuteTemplate(w, "verify", map[string]string{
				"Error":      "Invalid input.",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}

		switch *action {
		case "login":
			verifyLogin(w, r, *tokenCode)
		case "register":
			verifyRegister(w, r, *tokenCode)
		}

	})

	service.Route("POST", "/login", "handles login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var params siesta.Params
		loginEmail := params.String("login_email", "", "")
		loginPassword := params.String("login_password", "", "")
		password := params.Bool("password", false, "Use password based login")
		err := params.Parse(r.Form)
		if err != nil {
			log.Println(err)
			templ.ExecuteTemplate(w, "login_new", map[string]string{
				"Error":      "Something went wrong!",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}

		templateName := "login_new"
		if *password {
			templateName = "login_password"
		}

		user, err := MetadataClient.GetUserByEmail(*loginEmail)
		if err != nil {
			log.Println(err)
			templ.ExecuteTemplate(w, templateName, map[string]string{
				"Error":      "Something went wrong!",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}

		if !strings.HasSuffix(strings.ToLower(user.Email), "@preet.am") {
			templ.ExecuteTemplate(w, templateName, map[string]string{
				"Error":      "Sorry, Transverse is currently disabled for your user.",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}

		if *password && user.Verified {
			if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(*loginPassword)) == nil {
				userToken := &token.UserTokenData{
					User: user.ID,
				}
				encodedToken, err := TokenCodec.EncodeToken(token.NewToken(userToken, 0))
				if err != nil {
					log.Println(err)
					templ.ExecuteTemplate(w, templateName, map[string]string{
						"Error":      "Something went wrong!",
						"CDNDomain":  CDNDomain,
						"CDNVersion": CDNVersion,
					})
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     "transverse",
					Value:    encodedToken,
					Path:     "/",
					Expires:  time.Now().Add(7 * 24 * time.Hour),
					Secure:   !DevMode,
					HttpOnly: true,
				})

				w.Header().Set("Refresh", "2; /app")

				templ.ExecuteTemplate(w, "simple_message", map[string]string{
					"Success":    "Logged in! Taking you to the app...",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			} else {
				templ.ExecuteTemplate(w, templateName, map[string]string{
					"Error":      "Wrong password. If you forgot your password, login through email and change your password in your profile.",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
		}

		cookieValue := addSessionCookie(w, r, user.ID)

		codes := getTokenCodes(cookieValue, []byte(TokenKey), time.Now())

		if user.Verified {
			err = SendEmail(*loginEmail, "Transverse login token", fmt.Sprintf("Your login code is %06d.\nThis is only valid for a few minutes, so you better hurry!", codes[0]), CodeEmail("login", fmt.Sprintf("%06d", codes[0])))
			if err != nil {
				log.Println(err)
				templ.ExecuteTemplate(w, templateName, map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}

			w.Header().Set("Refresh", "0; /verify?action=login")
			return
		} else {
			if time.Now().Unix()-user.LastEmail < 3600 {
				// Already sent an email this hour
				w.WriteHeader(http.StatusBadRequest)
				templ.ExecuteTemplate(w, templateName, map[string]string{
					"Error":      "Rate limited. Please try again in an hour.",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
			err = SendEmail(*loginEmail, "Transverse registration token",
				fmt.Sprintf("Your registration code is %06d.\nThis is only valid for a few minutes, so you better hurry!", codes[0])+welcomePlaintext,
				CodeEmail("registration", fmt.Sprintf("%06d", codes[0]))+welcomeHTML,
			)
			if err != nil {
				log.Println(err)
				templ.ExecuteTemplate(w, templateName, map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}

			user.LastEmail = time.Now().Unix()
			MetadataClient.UpdateUser(user)

			w.Header().Set("Refresh", "0; /verify?action=register")
			return
		}

	})

	service.Route("POST", "/register", "handles register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// if true {
		// 	// REMOVE TO REACTIVATE REGISTRATION
		// 	templ.ExecuteTemplate(w, "register", map[string]string{
		// 		"Error":      "Registration is disabled for now.",
		// 		"CDNDomain":  CDNDomain,
		// 		"CDNVersion": CDNVersion,
		// 	})
		// 	return
		// }

		var params siesta.Params
		registerName := params.String("register_name", "", "")
		registerEmail := params.String("register_email", "", "")
		recaptchaResponse := params.String("g-recaptcha-response", "", "reCAPTCHA response")
		err := params.Parse(r.Form)
		if err != nil {
			log.Println(err)
			templ.ExecuteTemplate(w, "register", map[string]string{
				"Error":      "Something went wrong!",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}

		if !DevMode {
			// Check captcha
			if *recaptchaResponse == "" {
				w.WriteHeader(http.StatusBadRequest)
				templ.ExecuteTemplate(w, "register", map[string]string{
					"Error":      "Bad captcha",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}

			// verify CAPTCHA
			resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
				"secret":   []string{*recaptchaKey},
				"response": []string{*recaptchaResponse},
			})
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				templ.ExecuteTemplate(w, "register", map[string]string{
					"Error":      "Something went wrong! Please try again.",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
			recaptchaAPIResponse := struct {
				Success bool `json:"success"`
			}{}

			defer resp.Body.Close()
			err = json.NewDecoder(resp.Body).Decode(&recaptchaAPIResponse)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				templ.ExecuteTemplate(w, "register", map[string]string{
					"Error":      "Something went wrong! Please try again.",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
			if !recaptchaAPIResponse.Success {
				log.Println("recaptcha response:", recaptchaAPIResponse)
				w.WriteHeader(http.StatusBadRequest)
				templ.ExecuteTemplate(w, "register", map[string]string{
					"Error":      "Bad captcha",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
		}

		notFound := false
		user, err := MetadataClient.GetUserByEmail(*registerEmail)
		if err != nil {
			if serverErr, ok := err.(client.ServerError); ok && serverErr == http.StatusNotFound {
				notFound = true

				if *mgPublicKey != "" {
					verification, err := mg.ValidateEmail(*registerEmail)
					if err != nil {
						log.Println(err)
						templ.ExecuteTemplate(w, "register", map[string]string{
							"Error":      "Something went wrong!",
							"CDNDomain":  CDNDomain,
							"CDNVersion": CDNVersion,
						})
						return
					}
					if !verification.IsValid {
						templ.ExecuteTemplate(w, "register", map[string]string{
							"Error":      "Invalid email address",
							"CDNDomain":  CDNDomain,
							"CDNVersion": CDNVersion,
						})
						return
					}
				}

			} else {
				log.Println(err)
				templ.ExecuteTemplate(w, "register", map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
		}

		if notFound {
			// Create a user
			user = client.User{
				ID:        generateCode(8),
				Name:      *registerName,
				Email:     *registerEmail,
				Created:   time.Now().Unix(),
				Updated:   time.Now().Unix(),
				LastEmail: time.Now().Unix(),
			}
			err = MetadataClient.CreateUser(user)
			if err != nil {
				log.Println(err)
				templ.ExecuteTemplate(w, "register", map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
		} else {
			if time.Now().Unix()-user.LastEmail < 3600 {
				// Already sent an email this hour
				w.WriteHeader(http.StatusBadRequest)
				templ.ExecuteTemplate(w, "register", map[string]string{
					"Error":      "Rate limited. Please try again in an hour.",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
		}

		cookieValue := addSessionCookie(w, r, user.ID)

		codes := getTokenCodes(cookieValue, []byte(TokenKey), time.Now())
		err = SendEmail(*registerEmail, "Transverse registration token",
			fmt.Sprintf("Your registration code is %06d.\nThis is only valid for a few minutes, so you better hurry!", codes[0])+welcomePlaintext,
			CodeEmail("registration", fmt.Sprintf("%06d", codes[0]))+welcomeHTML,
		)
		if err != nil {
			log.Println(err)
			templ.ExecuteTemplate(w, "register", map[string]string{
				"Error":      "Something went wrong!",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}

		user.LastEmail = time.Now().Unix()
		MetadataClient.UpdateUser(user)

		w.Header().Set("Refresh", "0; /verify?action=register")
		return
	})

	service.Route("GET", "/logout", "serves logout page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.SetCookie(w, &http.Cookie{
			Name:     "transverse",
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(1 * time.Second),
			Secure:   !DevMode,
			HttpOnly: true,
		})

		w.Header().Set("Refresh", "2; /")

		templ.ExecuteTemplate(w, "simple_message", map[string]string{
			"Success":    "Logged out.",
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
	})

	service.Route("GET", "/app", "serves app page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		cookie, err := r.Cookie("transverse")
		if err != nil {
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		userTokenData := &token.UserTokenData{}
		_, err = TokenCodec.DecodeToken(cookie.Value, userTokenData)
		if err != nil {
			w.Header().Set("Location", "/login")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		templ.ExecuteTemplate(w, "app", map[string]string{
			"App":        "true",
			"Title":      "App",
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
	})

	service.SetNotFound(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		http.FileServer(http.Dir(*staticDir)).ServeHTTP(w, r)
	})
	log.Println("static directory set to", *staticDir)
	log.Println("listening on", *addr)

	var objectStore ObjectStore
	if *s3Key != "" {
		objectStore = &s3ObjectStore{s3: s3Service, bucket: "transverse"}
	} else {
		os.MkdirAll(*s3Directory, 0755)
		objectStore = &fileObjectStore{basePath: *s3Directory}
	}

	http.Handle(APIBasePath, NewAPI(objectStore).Service())
	http.Handle("/", service)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func addSessionCookie(w http.ResponseWriter, r *http.Request, userID string) string {
	log.Println("addSessionCookie")
	userToken := &token.UserTokenData{
		User: userID,
	}
	encodedToken, err := TokenCodec.EncodeToken(token.NewToken(userToken, 0))
	if err != nil {
		panic(err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "transverse_session",
		Value:    encodedToken,
		Path:     "/",
		Secure:   !DevMode,
		HttpOnly: true,
	})
	return encodedToken
}

func getTokenCodes(secret string, hmacKey []byte, timestamp time.Time) []int {
	codes := []int{}
	nowInt := int(timestamp.Unix()/300) * 300 // 5 min window
	for i := 0; i < 3; i++ {
		mac := hmac.New(sha256.New, hmacKey)
		mac.Write([]byte(secret + fmt.Sprint(nowInt-i*300)))
		sum := mac.Sum(nil)
		code := 0
		for j := 0; j < 4; j++ {
			code <<= 8
			code |= int(sum[j])
		}
		code = code % 999999
		codes = append(codes, code)
	}
	return codes
}

func verifyLogin(w http.ResponseWriter, r *http.Request, tokenCode int) {
	// Check session cookie
	cookie, err := r.Cookie("transverse_session")
	if err != nil {
		log.Println(err)
		w.Header().Set("Refresh", "2; /login")
		templ.ExecuteTemplate(w, "verify", map[string]string{
			"Error": "Invalid session.",
		})
		return
	}

	userTokenData := &token.UserTokenData{}
	_, err = TokenCodec.DecodeToken(cookie.Value, userTokenData)
	if err != nil {
		w.Header().Set("Refresh", "2; /login")
		templ.ExecuteTemplate(w, "verify", map[string]string{
			"Error": "Invalid session.",
		})
		return
	}

	codes := getTokenCodes(cookie.Value, []byte(TokenKey), time.Now())

	for _, code := range codes {
		if code == tokenCode {
			userToken := &token.UserTokenData{
				User: userTokenData.User,
			}
			encodedToken, err := TokenCodec.EncodeToken(token.NewToken(userToken, 0))
			if err != nil {
				log.Println(err)
				templ.ExecuteTemplate(w, "verify", map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "transverse",
				Value:    encodedToken,
				Path:     "/",
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				Secure:   !DevMode,
				HttpOnly: true,
			})

			w.Header().Set("Refresh", "2; /app")

			templ.ExecuteTemplate(w, "simple_message", map[string]string{
				"Success":    "Logged in! Taking you to the app...",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}
	}

	templ.ExecuteTemplate(w, "verify", map[string]string{
		"Error":      "Incorrect credentials.",
		"CDNDomain":  CDNDomain,
		"CDNVersion": CDNVersion,
	})
	return
}

func verifyRegister(w http.ResponseWriter, r *http.Request, tokenCode int) {
	// Check session cookie
	cookie, err := r.Cookie("transverse_session")
	if err != nil {
		log.Println(err)
		w.Header().Set("Refresh", "2; /register")
		templ.ExecuteTemplate(w, "verify", map[string]string{
			"Error":      "Invalid session.",
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
		return
	}

	userTokenData := &token.UserTokenData{}
	_, err = TokenCodec.DecodeToken(cookie.Value, userTokenData)
	if err != nil {
		w.Header().Set("Refresh", "2; /register")
		templ.ExecuteTemplate(w, "verify", map[string]string{
			"Error":      "Invalid session.",
			"CDNDomain":  CDNDomain,
			"CDNVersion": CDNVersion,
		})
		return
	}

	codes := getTokenCodes(cookie.Value, []byte(TokenKey), time.Now())

	for _, code := range codes {
		if code == tokenCode {
			user, err := MetadataClient.GetUserByID(userTokenData.User)
			if err != nil {
				log.Println(err)
				templ.ExecuteTemplate(w, "verify", map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}
			user.Verified = true
			err = MetadataClient.UpdateUser(user)
			if err != nil {
				log.Println(err)
				templ.ExecuteTemplate(w, "verify", map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}

			userToken := &token.UserTokenData{
				User: userTokenData.User,
			}

			encodedToken, err := TokenCodec.EncodeToken(token.NewToken(userToken, 0))
			if err != nil {
				log.Println(err)
				templ.ExecuteTemplate(w, "verify", map[string]string{
					"Error":      "Something went wrong!",
					"CDNDomain":  CDNDomain,
					"CDNVersion": CDNVersion,
				})
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "transverse",
				Value:    encodedToken,
				Path:     "/",
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				Secure:   !DevMode,
				HttpOnly: true,
			})

			w.Header().Set("Refresh", "2; /app")

			templ.ExecuteTemplate(w, "simple_message", map[string]string{
				"Success":    "Logged in! Taking you to the app...",
				"CDNDomain":  CDNDomain,
				"CDNVersion": CDNVersion,
			})
			return
		}
	}

	templ.ExecuteTemplate(w, "verify", map[string]string{
		"Error":      "Incorrect credentials.",
		"CDNDomain":  CDNDomain,
		"CDNVersion": CDNVersion,
	})
	return
}
