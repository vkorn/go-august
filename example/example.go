package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vkorn/go-august"
)

func main() {
	auth := august.NewAuthenticator(august.AuthMethPhone, "your-phone", "your-password", "", "gohomelock1")

	err := auth.Authenticate()

	if err != nil {
		fmt.Println("Authentication failed: " + err.Error())
		os.Exit(1)
	}

	if auth.State.State == august.AuthBadPassword {
		fmt.Println("Bad password")
		os.Exit(1)
	}

	if auth.State.State == august.AuthRequiresValidation {
		fmt.Println("Requesting verification code")
		err := auth.SendVerificationCode()
		if err != nil {
			fmt.Println("Verification failed: " + err.Error())
			os.Exit(1)
		}

		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		err = auth.ValidateCode(text)

		if err != nil {
			fmt.Println("Validation failed: " + err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully validated, calling auth once again")
		err = auth.Authenticate()
		if err != nil {
			fmt.Println("Final authentication failed: " + err.Error())
			os.Exit(1)
		}
	}

	fmt.Println("Authenticated, token is " + auth.State.AccessToken)

	api := august.NewAPIProvider(auth)
	l, err := api.GetLocks()
	if err != nil {
		fmt.Println("Failed to get locks: " + err.Error())
		os.Exit(1)
	}

	for _, v := range l {
		fmt.Println(fmt.Sprintf("Lock name: %s. Can operate: %t", v.Name, v.CanOperate()))
	}

	det, err := api.GetLockDetails(l[0].ID)
	if err != nil {
		fmt.Println("Failed to get lock details: " + err.Error())
		os.Exit(1)
	}

	fmt.Println(fmt.Sprintf("First lock ID: %s. Status: %v", det.ID, det.Status.Status))

	for _, v := range det.Users {
		fmt.Println(fmt.Sprintf("Lock user: %s. With type: %s. Can operate: %t",
			v.FirstName, v.UserType, v.CanOperate()))
	}

	st, err := api.GetLockStatus(l[0].ID)
	if err != nil {
		fmt.Println("Failed to get lock status: " + err.Error())
		os.Exit(1)
	}

	fmt.Println(fmt.Sprintf("Last status changed on: %s (%f seconds ago)",
		st.ChangedAt, st.SecondsSinceLastChange()))

	err = api.OpenLock(l[0].ID)

	if err != nil {
		fmt.Println("Failed to open lock: " + err.Error())
		os.Exit(1)
	}

	fmt.Println("Unlocked the lock")

	time.Sleep(5 * time.Second)
	err = api.CloseLock(l[0].ID)

	if err != nil {
		fmt.Println("Failed to close lock: " + err.Error())
		os.Exit(1)
	}

	fmt.Println("Locked the lock")
}
