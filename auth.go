package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	filename string
)

func AuthLoadFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	f.Close()

	filename = file
	return nil
}

func AuthReady() bool {
	return (filename != "")
}

func AuthFetch(username string) (string, string, error) {
	if !AuthReady() {
		return "", "", errors.New("Authentication file not specified. Call LoadFile() first")
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ':'
	reader.Comment = '#'

	re, _ := regexp.Compile("{([[:word:]]+)}")
	scheme := "BCRYPT"

	for {
		parts, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			return "", "", error
		}
		if (len(parts) > 1) && (strings.ToLower(username) == strings.ToLower(parts[0])) {
			m := re.FindStringSubmatch(parts[1])
			if len(m) == 2 {
				scheme = m[1]
				parts[1] = re.ReplaceAllString(parts[1], ``)
			}
			return parts[1], scheme, nil
		}
	}

	return "", "", errors.New("User not found")
}

func AuthCheckPassword(username string, secret string) error {
	hash, scheme, err := AuthFetch(username)
	if err != nil {
		return err
	}
	if strings.ToUpper(scheme) == "BCRYPT" {
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret)) == nil {
			return nil
		}
	} else if strings.ToUpper(scheme) == "PLAIN" {
		if hash == secret {
			return nil
		}
	} else {
		return errors.New("Unknown hashing scheme")
	}
	return errors.New("Password invalid")
}
