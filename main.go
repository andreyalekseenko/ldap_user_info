package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-ldap/ldap/v3"
)

func main() {
	http.HandleFunc("/user", getUserInfo)

	log.Fatal(http.ListenAndServe(":3333", nil))
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Username string `json:"username"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Error decoding JSON request body", http.StatusBadRequest)
		return
	}

	username := reqBody.Username
	ldapSearchBase := "OU=mag-rf.ru,DC=mag-rf,DC=ru"

	l, err := ldap.Dial("tcp", "dc1.mag-rf.ru:389")
	if err != nil {
		http.Error(w, "Error connecting to LDAP server", http.StatusInternalServerError)
		return
	}
	defer l.Close()

	err = l.Bind("cn=ldap_search2,dc=mag-rf,dc=ru", "a32008TYlk")
	if err != nil {
		http.Error(w, "Error binding to LDAP server", http.StatusInternalServerError)
		return
	}

	searchRequest := ldap.NewSearchRequest(
		ldapSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=user)(cn="+username+"))",
		[]string{"cn", "userPrincipalName", "ipPhone", "userAccountControl", "mail"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		http.Error(w, "Error searching LDAP server", http.StatusInternalServerError)
		return
	}

	if len(sr.Entries) == 0 {
		http.Error(w, "User not found in LDAP", http.StatusNotFound)
		return
	}
	//otsuda nachalo
	var userStatus string
	if sr.Entries[0].GetAttributeValue("userAccountControl") == "66048" || sr.Entries[0].GetAttributeValue("userAccountControl") == "512" {
		userStatus = "active"
	} else {
		userStatus = "disabled"
	}

	user := map[string]string{
		"firstName":     sr.Entries[0].GetAttributeValue("cn"),
		"login":         sr.Entries[0].GetAttributeValue("userPrincipalName"),
		"ipphone":       sr.Entries[0].GetAttributeValue("ipPhone"),
		"email":         sr.Entries[0].GetAttributeValue("mail"),
		"active_status": userStatus,
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Error encoding user data to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonData)
}
