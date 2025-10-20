package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/soderasen-au/go-common/crypto"
	"github.com/soderasen-au/go-common/util"
	"github.com/soderasen-au/go-qlik/qlik"
)

func usage() {
	prog := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s enc <user_domain> <user_id>    Encode JWT from user credentials\n", prog)
	fmt.Fprintf(os.Stderr, "  %s dec <jwt_token>                Decode and verify JWT\n", prog)
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])

	switch cmd {
	case "enc":
		if len(os.Args) < 4 {
			fatal("enc requires 2 arguments: user_domain and user_id")
		}

		claim := qlik.JwtClaim{
			UserID:        &os.Args[3],
			UserDirectory: &os.Args[2],
		}

		token, res := claim.GetJWT(crypto.InternalPrivateKey)
		if res != nil {
			fatal("%v", res)
		}

		fmt.Printf("\nPayload:\n%s\n\nJWT:\nBearer %s\n\n", util.Jsonify(claim.GetQlikClaims()), token)

	case "dec":
		claim := qlik.JwtClaim{}
		token := os.Args[2]
		verifyKey := &crypto.InternalPrivateKey.PublicKey

		_, err := jwt.ParseWithClaims(token, &claim, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return verifyKey, nil
		})
		if err != nil {
			fatal("failed to parse JWT: %v", err)
		}

		fmt.Printf("\nPayload:\n%s\n\nJWT:\n%s\n\n", util.Jsonify(&claim), token)

	default:
		fatal("invalid command '%s': must be 'enc' or 'dec'", cmd)
	}
}
