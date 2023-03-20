// package controllers

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"pubsubapi/env"
// 	"pubsubapi/errors"
// 	// "pubsubapi/internal/email"
// 	"pubsubapi/models"

// 	"github.com/gin-gonic/gin"
// 	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
// )

// const (

// 	// MAILTRAP config ----
// 	// defEmailHost        = "smtp.mailtrap.io"
// 	// defEmailPort        = "2525"
// 	// defEmailUsername    = "f1fb35bc899816"
// 	// defEmailPassword    = "b45ffc6e17857a"
// 	// defEmailFromAddress = "no-reply@discretal.com"
// 	// defEmailFromName    = "Discretal LOGIN Request"
// 	// defEmailTemplate    = "templates/smtp-notifier.tmpl"

// 	// OUTLOOK config ----
// 	// defEmailHost        = "smtp.office365.com"
// 	// defEmailPort        = "587"
// 	// defEmailUsername    = "w.debashree@leadergroup.com"
// 	// defEmailPassword    = "password"
// 	// defEmailFromAddress = "w.debashree@leadergroup.com" // this has to be same as defEmailUsername
// 	// defEmailFromName    = "Discretal LOGIN Request2"
// 	// defEmailTemplate    = "templates/smtp-notifier.tmpl"

// 	// GMAIL config ----
// 	defEmailHost        = "smtp.gmail.com"
// 	defEmailPort        = "587"
// 	defEmailUsername    = "doyalw@gmail.com"
// 	defEmailPassword    = "qpbfnimbyoevzknm"
// 	defEmailFromAddress = "no-reply@discretal.com"
// 	defEmailFromName    = "Discretal LOGIN Request"
// 	defEmailTemplate    = "templates/smtp-notifier.tmpl"

// 	defAdminEmail    = "admin@example.com"
// 	defAdminPassword = "12345678"

// 	envEmailHost        = "MF_EMAIL_HOST"
// 	envEmailPort        = "MF_EMAIL_PORT"
// 	envEmailUsername    = "MF_EMAIL_USERNAME"
// 	envEmailPassword    = "MF_EMAIL_PASSWORD"
// 	envEmailFromAddress = "MF_EMAIL_FROM_ADDRESS"
// 	envEmailFromName    = "MF_EMAIL_FROM_NAME"
// 	envEmailTemplate    = "MF_EMAIL_TEMPLATE"

// 	envAdminEmail    = "MF_USERS_ADMIN_EMAIL"
// 	envAdminPassword = "MF_USERS_ADMIN_PASSWORD"
// )

// // RegisterUser	godoc
// //
// //	@Summary		Registers user account
// //	@Description	Registers new user account given email and password. New account will be uniquely identified by its email address.
// //	@Tags			users
// //	@Produce		json
// //
// //	@Param			Request	body		models.RegisterUserReq	true	"JSON-formatted document describing the new user to be registered"
// //
// //	@Success		201		{object}	models.RegisterUserRes	"Registered new user."
// //	@Failure		400		"Failed due to malformed JSON."
// //	@Failure		500		"Unexpected server-side error occurred."
// //	@Router			/register [post]
// func RegisterUser(c *gin.Context) {
// 	var user models.RegisterUserReq
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	user.Status = "disabled" //disabled

// 	// check user already exists or not ->
// 	// configure the sdk payload for forwarding request
// 	urlUser := env.Env(envUserURL, sdkUserURL)
// 	configUser := sdk.Config{
// 		BaseURL: urlUser,
// 	}
// 	mfsdkUser := sdk.NewSDK(configUser)
// 	var sdkUser = sdk.User{
// 		Email:    env.Env(envAdminEmail, defAdminEmail),
// 		Password: env.Env(envAdminPassword, defAdminPassword),
// 	}
// 	token, err := mfsdkUser.CreateToken(sdkUser)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	req, _ := http.NewRequest("GET", urlUser+"/users", nil)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
// 	req.Header.Add("Content-Type", "application/json")

// 	type users struct {
// 		ID    string `json:"id"`
// 		Email string `json:"email"`
// 	}
// 	type respUsers struct {
// 		Users []users `json:"users"`
// 	}

// 	// "enabled" status
// 	//add query parameters
// 	q := req.URL.Query()
// 	q.Add("email", user.Email)
// 	q.Add("status", "enabled")
// 	req.URL.RawQuery = q.Encode()

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	data, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	var rpUsers respUsers
// 	json.Unmarshal(data, &rpUsers)

// 	if len(rpUsers.Users) > 0 { // means email already exists
// 		errors.InfoHandler(c, "Email ID already registered. Please login using the same.", "Email ID already registered. Please login using the same.", http.StatusAccepted)
// 		return
// 	}

// 	// "disabled" status
// 	q.Del("status")
// 	q.Add("status", "disabled")
// 	req.URL.RawQuery = q.Encode()

// 	resp, err = client.Do(req)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	data, err = io.ReadAll(resp.Body)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	json.Unmarshal(data, &rpUsers)

// 	if len(rpUsers.Users) > 0 { // means email already exists
// 		errors.InfoHandler(c, "Registration request already received. Verification is in progress.", "Registration request already received. Verification is in progress.", http.StatusOK)
// 		return
// 	}

// 	// "flagged" status
// 	q.Del("status")
// 	q.Add("status", "flagged")
// 	req.URL.RawQuery = q.Encode()

// 	resp, err = client.Do(req)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	data, err = io.ReadAll(resp.Body)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	json.Unmarshal(data, &rpUsers)

// 	if len(rpUsers.Users) > 0 { // means email already exists
// 		errors.InfoHandler(c, "Registration request denied. Please contact admin for any queries.", "Registration request denied. Please contact admin for any queries.", http.StatusOK)
// 		return
// 	}

// 	// configure the sdk payload for forwarding request
// 	registrationURL := urlUser + "/users"
// 	body, _ := json.Marshal(user)
// 	regisreq, err := http.NewRequest("POST", registrationURL, bytes.NewBuffer(body))
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
// 		return
// 	}
// 	regisreq.Header.Set("Content-Type", "application/json")

// 	regisresp, err := client.Do(regisreq)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
// 	}
// 	defer regisresp.Body.Close()

// 	var result string
// 	code := regisresp.StatusCode
// 	fmt.Println("->>", code)

// 	// // configuration for email
// 	// cfg := email.Config{
// 	// 	Host:        env.Env(envEmailHost, defEmailHost),
// 	// 	Port:        env.Env(envEmailPort, defEmailPort),
// 	// 	Username:    env.Env(envEmailUsername, defEmailUsername),
// 	// 	Password:    env.Env(envEmailPassword, defEmailPassword),
// 	// 	FromAddress: env.Env(envEmailFromAddress, defEmailFromAddress),
// 	// 	FromName:    env.Env(envEmailFromName, defEmailFromName),
// 	// 	Template:    env.Env(envEmailTemplate, defEmailTemplate),
// 	// }

// 	if code == 200 || code == 201 {
// 		result = "registration request submitted successfully"
// 		// ag, err := email.New(&cfg)
// 		// if err != nil {
// 		// 	errors.ErrHandler(c, fmt.Errorf("failed to configure e-mailing util : %v", err), http.StatusBadRequest)
// 		// 	return
// 		// }

// 		// // compose the email's content here - header, footer, etc.
// 		// err = ag.Send([]string{user.Email}, "", "Registration request received", "", "Request is receieved.", "")
// 		// if err != nil {
// 		// 	errors.ErrHandler(c, fmt.Errorf("error occurred while sending email : %v", err), http.StatusBadRequest)
// 		// 	return
// 		// }
// 	} else {
// 		result = "registration unsuccessful - please try again later"
// 	}

// 	// errors.InfoHandler(c, "user registered successfully", userID, http.StatusCreated)
// 	errors.InfoHandler(c, result, result, code)
// }

// func ListUnverifiedUsers(c *gin.Context) {
// 	// check if auth token exists
// 	token, err := getToken(c)
// 	if err != nil {
// 		errors.ErrHandler(c, err, http.StatusUnauthorized)
// 		return
// 	}

// 	urlUser := env.Env(envUserURL, sdkUserURL)
// 	req, _ := http.NewRequest("GET", urlUser+"/users", nil)
// 	req.Header.Add("Authorization", token)
// 	req.Header.Add("Content-Type", "application/json")

// 	type users struct {
// 		ID    string `json:"id"`
// 		Email string `json:"email"`
// 	}
// 	type respUsers struct {
// 		Users []users `json:"users"`
// 	}

// 	// "enabled" status
// 	//add query parameters
// 	q := req.URL.Query()
// 	q.Add("status", "disabled")
// 	req.URL.RawQuery = q.Encode()

// 	fmt.Println(req.URL.RawQuery)
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	fmt.Println(resp.StatusCode)
// 	data, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	var rpUsers respUsers
// 	json.Unmarshal(data, &rpUsers)

// 	errors.InfoHandler(c, "unverified users list generated successfully", rpUsers, http.StatusOK)
// }

// func RespondUnverifiedUsers(c *gin.Context) {
// 	// check if auth token exists
// 	token, err := getToken(c)
// 	if err != nil {
// 		errors.ErrHandler(c, err, http.StatusUnauthorized)
// 		return
// 	}

// 	type VerifyUser struct {
// 		ID     string `json:"id" binding:"required"`
// 		Action string `json:"action" binding:"required"`
// 	}
// 	var verifyUser VerifyUser
// 	if err := c.ShouldBindJSON(&verifyUser); err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	if verifyUser.Action != "enable" &&
// 		verifyUser.Action != "disable" &&
// 		verifyUser.Action != "flag" {
// 		errors.ErrHandler(c, fmt.Errorf("invalid action selected"), http.StatusBadRequest)
// 		return
// 	}

// 	urlUser := env.Env(envUserURL, sdkUserURL)
// 	newURL := urlUser + "/users/" + verifyUser.ID + "/" + verifyUser.Action
// 	req, _ := http.NewRequest("POST", newURL, nil)
// 	req.Header.Add("Authorization", token)
// 	req.Header.Add("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	fmt.Println(resp.StatusCode)

// 	errors.InfoHandler(c, "Verification done", "Verification done", http.StatusOK)
// }

//=========================================================================

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pubsubapi/env"
	"pubsubapi/errors"
	"pubsubapi/internal/email"
	"pubsubapi/models"

	"github.com/gin-gonic/gin"
	// sdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

const (

	// MAILTRAP config ----
	// defEmailHost        = "smtp.mailtrap.io"
	// defEmailPort        = "2525"
	// defEmailUsername    = "f1fb35bc899816"
	// defEmailPassword    = "b45ffc6e17857a"
	// defEmailFromAddress = "no-reply@discretal.com"
	// defEmailFromName    = "Discretal LOGIN Request"
	// defEmailTemplate    = "templates/smtp-notifier.tmpl"

	// OUTLOOK config ----
	// defEmailHost        = "smtp.office365.com"
	// defEmailPort        = "587"
	// defEmailUsername    = "w.debashree@leadergroup.com"
	// defEmailPassword    = "password"
	// defEmailFromAddress = "w.debashree@leadergroup.com" // this has to be same as defEmailUsername
	// defEmailFromName    = "Discretal LOGIN Request2"
	// defEmailTemplate    = "templates/smtp-notifier.tmpl"

	// GMAIL config ----
	defEmailHost     = "smtp.gmail.com"
	defEmailPort     = "587"
	defEmailUsername = "discretalfms@gmail.com"
	defEmailPassword = "oatxlpjaszgdibdb"
	// defEmailPassword    = "qpbfnimbyoevzknm"
	defEmailFromAddress = "no-reply@discretal.com"
	defEmailFromName    = "Discretal Login Request"
	// defEmailTemplate    = "templates/registration.tmpl"
	defEmailTemplate = "registration.tmpl"

	defAdminEmail        = ""
	defAdminPassword     = ""
	defTestAdminEmail    = "doyal03@gmail.com"
	defTestAdminPassword = "12345678"

	envEmailHost        = "MF_EMAIL_HOST"
	envEmailPort        = "MF_EMAIL_PORT"
	envEmailUsername    = "MF_EMAIL_USERNAME"
	envEmailPassword    = "MF_EMAIL_PASSWORD"
	envEmailFromAddress = "MF_EMAIL_FROM_ADDRESS"
	envEmailFromName    = "MF_EMAIL_FROM_NAME"
	envEmailTemplate    = "MF_REGISTRATION_EMAIL_TEMPLATE"

	envAdminEmail        = "MF_USERS_ADMIN_EMAIL"
	envAdminPassword     = "MF_USERS_ADMIN_PASSWORD"
	envTestAdminEmail    = "MF_USERS_TEST_ADMIN_EMAIL"
	envTestAdminPassword = "MF_USERS_TEST_ADMIN_PASSWORD"
)

// RegisterUser	godoc
//
//	@Summary		Registers user account
//	@Description	Registers new user account given email and password. New account will be uniquely identified by its email address.
//	@Tags			users
//	@Produce		json
//
//	@Param			Request	body		models.RegisterUserReq	true	"JSON-formatted document describing the new user to be registered"
//
//	@Success		201		{object}	models.RegisterUserRes	"Registered new user."
//	@Failure		400		"Failed due to malformed JSON."
//	@Failure		500		"Unexpected server-side error occurred."
//	@Router			/register [post]
func RegisterUser(c *gin.Context) {
	var newUser models.RegisterUserReq
	if err := c.ShouldBindJSON(&newUser); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	type UserReq struct {
		Email    string          `json:"email" binding:"required"`
		Password string          `json:"password" binding:"required"`
		Status   string          `json:"status" binding:"required"`
		Metadata models.Metadata `json:"metadata,omitempty"`
	}
	user := UserReq{
		Email:    newUser.Email,
		Password: newUser.Password,
		Status:   "disabled",
		Metadata: map[string]interface{}{
			"profile": map[string]string{
				"firstName": newUser.Firstname,
				"lastName":  newUser.Lastname,
				"email":     newUser.Email,
				"picture":   "assets/images/mainflux-logo.png",
			},
		},
	}
	// user.Status = "disabled" //disabled

	// check user already exists or not ->
	// configure the sdk payload for forwarding request
	urlUser := env.Env(envUserURL, sdkUserURL)
	// configUser := sdk.Config{
	// 	BaseURL: urlUser,
	// }
	// mfsdkUser := sdk.NewSDK(configUser)

	var admin = models.LoginUserReq{
		Email:    env.Env(envAdminEmail, defAdminEmail),
		Password: env.Env(envAdminPassword, defAdminPassword),
	}
	// token, err := mfsdkUser.CreateToken(sdkUser)
	// if err != nil {
	// 	errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
	// 	return
	// }

	// check user already exists or not ->

	body, _ := json.Marshal(admin)
	fmt.Println(string(body))
	tokenreq, err := http.NewRequest("POST", urlUser+"/tokens", bytes.NewBuffer(body))
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	tokenreq.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	tokenresp, err := client.Do(tokenreq)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer tokenresp.Body.Close()

	data, err := io.ReadAll(tokenresp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	type TokenRes struct {
		Token string `json:"token,omitempty"`
	}
	var tokenRes TokenRes
	json.Unmarshal(data, &tokenRes)
	token := tokenRes.Token
	fmt.Println("token=>", token)

	req, _ := http.NewRequest("GET", urlUser+"/users", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	type users struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	type respUsers struct {
		Users []users `json:"users"`
	}

	// "enabled" status
	//add query parameters
	q := req.URL.Query()
	q.Add("email", user.Email)
	q.Add("status", "enabled")
	req.URL.RawQuery = q.Encode()

	// client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	var rpUsers respUsers
	json.Unmarshal(data, &rpUsers)

	if len(rpUsers.Users) > 0 { // means email already exists
		errors.InfoHandler(c, "Email ID already registered. Please login using the same.", "Email ID already registered. Please login using the same.", http.StatusAccepted)
		return
	}

	// "disabled" status
	q.Del("status")
	q.Add("status", "disabled")
	req.URL.RawQuery = q.Encode()

	resp, err = client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	json.Unmarshal(data, &rpUsers)

	if len(rpUsers.Users) > 0 { // means email already exists
		errors.InfoHandler(c, "Registration request already received. Verification is in progress.", "Registration request already received. Verification is in progress.", http.StatusOK)
		return
	}

	// "flagged" status
	q.Del("status")
	q.Add("status", "flagged")
	req.URL.RawQuery = q.Encode()

	resp, err = client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	json.Unmarshal(data, &rpUsers)

	if len(rpUsers.Users) > 0 { // means email already exists
		errors.InfoHandler(c, "Registration request denied. Please contact admin for any queries.", "Registration request denied. Please contact admin for any queries.", http.StatusOK)
		return
	}

	// configure the sdk payload for forwarding request
	registrationURL := urlUser + "/users"
	body, _ = json.Marshal(user)
	regisreq, err := http.NewRequest("POST", registrationURL, bytes.NewBuffer(body))
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
		return
	}
	regisreq.Header.Set("Content-Type", "application/json")

	regisresp, err := client.Do(regisreq)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusInternalServerError)
	}
	defer regisresp.Body.Close()

	var result string
	code := regisresp.StatusCode
	if code == 200 || code == 201 {
		result = "registration request submitted successfully"
	} else {
		result = "registration unsuccessful - please try again later"
	}

	go func() {
		// configuration for email
		cfg := email.Config{
			Host:        env.Env(envEmailHost, defEmailHost),
			Port:        env.Env(envEmailPort, defEmailPort),
			Username:    env.Env(envEmailUsername, defEmailUsername),
			Password:    env.Env(envEmailPassword, defEmailPassword),
			FromAddress: env.Env(envEmailFromAddress, defEmailFromAddress),
			FromName:    env.Env(envEmailFromName, defEmailFromName),
			Template:    env.Env(envEmailTemplate, defEmailTemplate),
		}

		if code == 200 || code == 201 {
			ag, err := email.New(&cfg)
			if err != nil {
				errors.ErrHandler(c, fmt.Errorf("failed to configure e-mailing util : %v", err), http.StatusBadRequest)
				return
			}

			// compose the email's content here - header, footer, etc.
			var Receivers []string
			if val := env.Env(envAdminEmail, defAdminEmail); val != "" {
				Receivers = append(Receivers, val)
			}
			if val := env.Env(envTestAdminEmail, defTestAdminEmail); val != "" {
				Receivers = append(Receivers, val)
			}
			// Receivers = append(Receivers, user.Email)

			// 	rawContent := []byte(`Hello %s,

			// Thank you for registering on our Discretal platform. You will be able to login to the platform once your request has been approved.`)
			// 	content := fmt.Sprintf(string(rawContent), newUser.Name)

			// err = ag.Send(user.Email, Receivers, "", "New registration request received", "", "New registration request is receieved  from "+user.Email+".", "")
			// err = ag.Send(user.Email, Receivers, "", "Discretal registration request received", "", content, "")
			err = ag.Send(user.Email, Receivers, "", "Discretal registration request received", "Dear "+newUser.Firstname+" "+newUser.Lastname, "", "")
			if err != nil {
				errors.ErrHandler(c, fmt.Errorf("error occurred while sending email : %v", err), http.StatusBadRequest)
				return
			}
		}
	}()

	// errors.InfoHandler(c, "user registered successfully", userID, http.StatusCreated)
	errors.InfoHandler(c, result, result, code)
}

func ListUnverifiedUsers(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	urlUser := env.Env(envUserURL, sdkUserURL)
	req, _ := http.NewRequest("GET", urlUser+"/users", nil)
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	type pageRes struct {
		Total  uint64 `json:"total"`
		Offset uint64 `json:"offset"`
		Limit  uint64 `json:"limit"`
	}
	type users struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	type respUsers struct {
		pageRes
		Users []users `json:"users"`
	}

	// "enabled" status
	//add query parameters
	q := req.URL.Query()
	q.Add("status", "disabled")
	params := []string{"offset", "limit"}
	for _, prms := range params {
		val, ok := c.GetQuery(prms)
		if ok {
			q.Add(prms, val)
		}
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	var rpUsers respUsers
	json.Unmarshal(data, &rpUsers)

	errors.InfoHandler(c, "unverified users list generated successfully", rpUsers, http.StatusOK)
}

func RespondUnverifiedUsers(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	type VerifyUser struct {
		ID     string `json:"id" binding:"required"`
		Action string `json:"action" binding:"required"`
	}
	var verifyUsers []VerifyUser
	if err := c.ShouldBindJSON(&verifyUsers); err != nil {
		errors.ErrHandler(c, fmt.Errorf("error at binding values : %v", err), http.StatusBadRequest)
		return
	}

	for _, verifyUser := range verifyUsers {

		// if verifyUser.Action != "enable" &&
		// 	verifyUser.Action != "disable" &&
		// 	verifyUser.Action != "flag" {
		// 	errors.ErrHandler(c, fmt.Errorf("invalid action selected"), http.StatusBadRequest)
		// 	return
		// }
		if verifyUser.Action != "approve" &&
			verifyUser.Action != "reject" {
			errors.ErrHandler(c, fmt.Errorf("invalid action selected"), http.StatusBadRequest)
			return
		}

		var action string
		if verifyUser.Action == "approve" {
			action = "enable"
		} else {
			action = "flag"
		}

		urlUser := env.Env(envUserURL, sdkUserURL)
		// newURL := urlUser + "/users/" + verifyUser.ID + "/" + verifyUser.Action
		newURL := urlUser + "/users/" + verifyUser.ID + "/" + action
		req, _ := http.NewRequest("POST", newURL, nil)
		req.Header.Add("Authorization", token)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()
	}

	errors.InfoHandler(c, "Verification done", "Verification done", http.StatusOK)
}

func RejectedUsers(c *gin.Context) {
	// check if auth token exists
	token, err := getToken(c)
	if err != nil {
		errors.ErrHandler(c, err, http.StatusUnauthorized)
		return
	}

	urlUser := env.Env(envUserURL, sdkUserURL)
	req, _ := http.NewRequest("GET", urlUser+"/users", nil)
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	type pageRes struct {
		Total  uint64 `json:"total"`
		Offset uint64 `json:"offset"`
		Limit  uint64 `json:"limit"`
	}
	type users struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	type respUsers struct {
		pageRes
		Users []users `json:"users"`
	}

	//add query parameters
	q := req.URL.Query()
	q.Add("status", "flagged")
	params := []string{"offset", "limit"}
	for _, prms := range params {
		val, ok := c.GetQuery(prms)
		if ok {
			q.Add(prms, val)
		}
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("failed to get response : %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ErrHandler(c, fmt.Errorf("request could not be processed by server : %v", err), http.StatusBadRequest)
		return
	}

	var rpUsers respUsers
	json.Unmarshal(data, &rpUsers)

	errors.InfoHandler(c, "unverified users list generated successfully", rpUsers, http.StatusOK)
}
