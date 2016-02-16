package passwd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type User struct {
	ID       int
	Username string
}

type Users []User

func (users Users) NameForID(id int) (string, bool) {
	for _, user := range users {
		if id == user.ID {
			return user.Username, true
		}
	}
	return "", false
}

func ReadUsers(path string) (Users, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	userScanner := bufio.NewScanner(file)

	users := []User{}
	lineCount := 0
	for userScanner.Scan() {
		lineCount++
		userLine := strings.TrimSpace(userScanner.Text())
		if userLine == "" || strings.HasPrefix(userLine, "#") {
			continue
		}
		userLineColumns := strings.Split(userLine, ":")
		if len(userLineColumns) != 7 {
			return nil, fmt.Errorf("malformed user on line %d", lineCount)
		}
		userName := userLineColumns[0]
		userIDStr := userLineColumns[2]
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("malformed user ID on line %d: %s", lineCount, userIDStr)
		}
		users = append(users, User{
			Username: userName,
			ID:       userID,
		})
	}
	return users, nil
}
