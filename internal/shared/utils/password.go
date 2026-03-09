package utils

import "golang.org/x/crypto/bcrypt"

func HashPwd(plainPwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPwd), bcrypt.DefaultCost)
	return string(hash), err
}

func ComparePwd(hashPwd, plainPwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashPwd), []byte(plainPwd))
	return err
}
