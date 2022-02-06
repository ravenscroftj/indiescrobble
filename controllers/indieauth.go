package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"git.jamesravey.me/ravenscroftj/indiescrobble/config"
	"git.jamesravey.me/ravenscroftj/indiescrobble/models"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"
	"github.com/hacdias/indieauth"
	"github.com/lestrrat-go/jwx/jwt"
)

type IndieAuthManager struct {
	iac *indieauth.Client
	jwtAuth *jwtauth.JWTAuth
}

func NewIndieAuthManager() *IndieAuthManager{
	
	config := config.GetConfig()

	iam := new(IndieAuthManager)
	iam.iac = indieauth.NewClient(config.GetString("indieauth.clientName"), config.GetString("indieauth.redirectURL"), nil)


	iam.jwtAuth = jwtauth.New("HS256", []byte(config.GetString("jwt.signKey")), []byte(config.GetString("jwt.signKey")))

	return iam
}

func (iam *IndieAuthManager) GetCurrentUser(c *gin.Context) *models.BaseUser {

	jwt, err := c.Cookie("jwt")

	if err != nil {
		return nil
	}else{
		tok, err := iam.jwtAuth.Decode(jwt)

		if err != nil{
			log.Printf("Failed to decode jwt: %v", err)
			return nil
		}

		me, present := tok.Get("user")

		if !present{
			return nil
		}

		indietok, present := tok.Get("token")

		if !present{
			return nil
		}

		user := models.BaseUser{Me:  me.(string), Token: indietok.(string)}

		return &user

	}

}

func (iam *IndieAuthManager) getInformation(c *gin.Context) (*indieauth.AuthInfo, string, error) {


	config := config.GetConfig()

	cookie, err := c.Request.Cookie(config.GetString("indieauth.oauthCookieName"))
	if err != nil {
		return nil, "", err
	}

	token, err := jwtauth.VerifyToken(iam.jwtAuth, cookie.Value)
	if err != nil {
		return nil, "", err
	}

	err = jwt.Validate(token)
	if err != nil {
		return nil, "", err
	}

	if token.Subject() != config.GetString("indieauth.oauthSubject") {
		return nil, "", errors.New("invalid subject for oauth token")
	}

	data, ok := token.Get("data")
	if !ok || data == nil {
		return nil, "", errors.New("cannot find 'data' property in token")
	}

	dataStr, ok := data.(string)
	if !ok || dataStr == "" {
		return nil, "", errors.New("cannot find 'data' property in token")
	}

	var i *indieauth.AuthInfo
	err = json.Unmarshal([]byte(dataStr), &i)
	if err != nil {
		return nil, "", err
	}

	// Delete cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     config.GetString("indieauth.oauthCookieName"),
		MaxAge:   -1,
		Secure:   c.Request.URL.Scheme == "https",
		Path:     "/",
		HttpOnly: true,
	})

	redirect, ok := token.Get("redirect")
	if !ok {
		return i, "", nil
	}

	redirectStr, ok := redirect.(string)
	if !ok || redirectStr == "" {
		return i, "", nil
	}

	return i, redirectStr, nil
}

func (iam *IndieAuthManager) saveAuthInfo(w http.ResponseWriter, r *http.Request, i *indieauth.AuthInfo) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}

	config := config.GetConfig()

	expiration := time.Now().Add(time.Minute * 10)

	_, signed, err := iam.jwtAuth.Encode(map[string]interface{}{
		jwt.SubjectKey:    config.GetString("indieauth.oauthSubject"),
		jwt.IssuedAtKey:   time.Now().Unix(),
		jwt.ExpirationKey: expiration,
		"data":            string(data),
		"redirect":        r.URL.Query().Get("redirect"),
	})
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     config.GetString("indieauth.oauthCookieName"),
		Value:    string(signed),
		Expires:  expiration,
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	return nil
}


func (iam *IndieAuthManager) Logout(c *gin.Context) {

	// delete the cookie
	cookie := &http.Cookie{
		Name:     "jwt",
		MaxAge:   -1,
		Secure:   c.Request.URL.Scheme == "https",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(c.Writer, cookie)
	
	// redirect back to index
	c.Redirect(http.StatusSeeOther, "/")

}

func (iam *IndieAuthManager) IndieAuthLoginPost(c *gin.Context) {

	err := c.Request.ParseForm()

	if err != nil{
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	r := c.Request

	profile := r.FormValue("domain")
	if profile == "" {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": "Empty domain",
		})
		return
	}

	profile = indieauth.CanonicalizeURL(profile)
	if err := indieauth.IsValidProfileURL(profile); err != nil {
		err = fmt.Errorf("invalid profile url: %w", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	i, redirect, err := iam.iac.Authenticate(profile, "")

	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	fmt.Printf("profile: %v\n", i)


	err = iam.saveAuthInfo(c.Writer, c.Request, i)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	// append me param so the user doesn't have to enter this twice
	redirect = fmt.Sprintf("%v&me=%v", redirect, url.QueryEscape(i.Me) )

	c.Redirect(http.StatusSeeOther, redirect)
}


func (iam *IndieAuthManager) LoginCallbackGet(c *gin.Context) {

	config := config.GetConfig()

	i, redirect, err := iam.getInformation(c)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	code, err := iam.iac.ValidateCallback(i, c.Request)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}


	// profile, err := iam.iac.FetchProfile(i, code)
	// if err != nil {
	// 	c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
	// 		"message": err,
	// 	})
	// 	return
	// }


	token, _, err := iam.iac.GetToken(i, code)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	me := token.Extra("me").(string)


	if err := indieauth.IsValidProfileURL(me); err != nil {
		err = fmt.Errorf("invalid 'me': %w", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	expiration := time.Now().Add(time.Hour * 24 * 7)

	_, signed, err := iam.jwtAuth.Encode(map[string]interface{}{
		jwt.SubjectKey:    config.GetString("indieauth.sessionSubject"),
		jwt.IssuedAtKey:   time.Now().Unix(),
		jwt.ExpirationKey: expiration,
		"user":            me,
		"token": token.AccessToken,
	})
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"message": err,
		})
		return
	}

	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    string(signed),
		Expires:  expiration,
		Secure:   c.Request.URL.Scheme == "https",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(c.Writer, cookie)

	if redirect == "" {
		redirect = "/"
	}

	c.Redirect(http.StatusSeeOther, redirect)
}