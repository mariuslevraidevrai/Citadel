package main

import (
	"os"
	"strings"
	"log"

	"golang.org/x/crypto/bcrypt"
)

const pwdFile = "password.key"

func setPassword(plainPassword string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(plainPassword)), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    return os.WriteFile(pwdFile, hash, 0600)
}

func initAuth() error {
	if _, err := os.Stat(pwdFile); os.IsNotExist(err) {
		defaultPassword := "citadel"
		hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		return os.WriteFile(pwdFile, hash, 0600)
	}
	return nil
}

func checkPassword(plainPassword string) (bool, error) {
	hashBytes, err := os.ReadFile(pwdFile)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword(hashBytes, []byte(strings.TrimSpace(plainPassword)))
	if err != nil {
		log.Printf("Erreur bcrypt detail : %v", err) // Permet de voir si c'est un problème de hachage
		return false, nil
	}
	return true, nil
	
}