package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

type Config struct {
	idpMetadataURL string

	ServiceProviderId string
	spRootURL         string

	certFile    string
	certKeyFile string
}

func (c *Config) getConfig() {
	if env_var := os.Getenv("SAMLRQT_IDP_URLMETADATA"); env_var != "" {
		c.idpMetadataURL = env_var
	} else {
		log.Fatalf("IDP Metadata URL not defined ... \nNeed to create environment var : SAMLRQT_IDP_URLMETADATA")
	}

	if env_var := os.Getenv("SAMLRQT_SP_ROOTURL"); env_var != "" {
		c.spRootURL = env_var
	} else {
		log.Fatalf("SP Root URL not defined ... \nNeed to create environment var : SAMLRQT_SP_ROOTURL")
	}

	if env_var := os.Getenv("SAMLRQT_SP_ID"); env_var != "" {
		c.ServiceProviderId = env_var
	} else {
		log.Fatalf("SP IndentityID not defined ... \nNeed to create environment var : SAMLRQT_SP_ID")
	}

	if env_var := os.Getenv("SAMLRQT_CERT_FILE"); env_var != "" {
		c.certFile = env_var
	} else {
		log.Fatalf("Certificate path not defined ... \nNeed to create environment var : SAMLRQT_CERT_FILE")
	}

	if env_var := os.Getenv("SAMLRQT_CERT_KEY"); env_var != "" {
		c.certKeyFile = env_var
	} else {
		log.Fatalf("Certificate key path not defined ... \nNeed to create environment var : SAMLRQT_CERT_KEY")
	}

	os.Stdout.WriteString("Loaded parameters :\n")
	os.Stdout.WriteString("- SAMLRQT_IDP_URLMETADATA :" + c.idpMetadataURL + "\n")
	os.Stdout.WriteString("- SAMLRQT_SP_ROOTURL :" + c.spRootURL + "\n")
	os.Stdout.WriteString("- SAMLRQT_SP_ID :" + c.ServiceProviderId + "\n")
	os.Stdout.WriteString("- SAMLRQT_CERT_FILE :" + c.certFile + "\n")
	os.Stdout.WriteString("- SAMLRQT_CERT_KEY :" + c.certKeyFile + "\n")
	os.Stdout.WriteString("\n")
}

func displayRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dans handlerSamlRequest")

	io.WriteString(w, "You're under SAML-REQUESTER"+"\n")
	io.WriteString(w, "  - Method : "+r.Method+"\n")
	io.WriteString(w, "  - Headers : "+r.Header.Get("User-Agent")+"\n")

	fmt.Println("request")
	fmt.Println(r)
	fmt.Println("request.Header")
	fmt.Println(r.Header)
	for _, cookie := range r.Cookies() {
		fmt.Println("Found a cookie named:", cookie.Name)
		io.WriteString(w, "<b>get cookie"+cookie.Name+" value is "+cookie.Value+"</b>\n")
	}
}

func DisplayIDPMetadata(m *saml.EntityDescriptor) {
	fmt.Println("idpMetadataContent :")
	fmt.Println("IDP ID :" + m.ID)
	fmt.Println("IDP EntityID :" + m.EntityID)

	//	fmt.Println(m.IDPSSODescriptors )
	//	fmt.Println(m.AuthnAuthorityDescriptors)
}

func samlRequester(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dans samlRequester")
	// Ce qu'il va renvoyer :
	fmt.Println(samlsp.AttributeFromContext(r.Context(), "cn"))
	for _, cookie := range r.Cookies() {
		fmt.Println("Found a cookie named:", cookie.Name)
	}
	fmt.Fprintf(w, "SAMLRequester, %s!", samlsp.AttributeFromContext(r.Context(), "cn"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func main() {
	cfg := new(Config)
	cfg.getConfig()

	fmt.Println("Initialize Certificate KeyPair Signature ...")
	keyPair, err := tls.LoadX509KeyPair(cfg.certFile, cfg.certKeyFile)
	if err != nil {
		log.Fatalf("failed to read Signature certification files: %v", err)
		panic(err) // TODO handle error
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		panic(err) // TODO handle error
	}

	fmt.Println("Retrieve idpMetadataURL from Config ...")
	idpMetadataURL, err := url.Parse(cfg.idpMetadataURL)
	if err != nil {
		panic(err) // TODO handle error
	}
	fmt.Println(idpMetadataURL)

	fmt.Println("Retrieve SPRootURL from Config ...")
	rootURL, err := url.Parse(cfg.spRootURL)
	if err != nil {
		panic(err) // TODO handle error
	}
	//	fmt.Println("rootURL")
	//	fmt.Println(rootURL)

	fmt.Println("Fetch Metadata From IDP metadataURL...")
	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		panic(err) // TODO handle error
	}
	//	DisplayIDPMetadata(idpMetadata)

	fmt.Println("Create SAML AuthnRequest")
	samlSP, _ := samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
		EntityID:    cfg.ServiceProviderId,
		SignRequest: true,
	})
	//	fmt.Println("Value au New")
	//	fmt.Println(samlSP.ServiceProvider.AcsURL)
	//	fmt.Println(samlSP.ServiceProvider.EntityID)

	// I need to revaluate the AcsURL value because under the New,
	// It add automatically acs at the end of the URL and delete the path
	samlSP.ServiceProvider.AcsURL = *rootURL

	//	fmt.Println("Value apres revalorisation")
	//	fmt.Println(samlSP.ServiceProvider.AcsURL)
	//	fmt.Println(samlSP.ServiceProvider.EntityID)

	//	fmt.Println("Service PRovider details...")
	//	fmt.Println(samlSP.ServiceProvider)

	//	os.Stdout.WriteString("- IDP Metadata - IDP AuthnRequest :" + samlSP.idpMetadataURL.AcsURL.String() + "\n")

	app := http.HandlerFunc(samlRequester)
	//	fmt.Println(app)
	http.Handle("/saml-requester", samlSP.RequireAccount(app))
	http.HandleFunc("/request", displayRequest)
	http.Handle("/saml/", samlSP)
	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		// handle your error here
		log.Fatalf("HTTP server fail : %v", err)
		panic(err) // TODO handle error
	}
}
