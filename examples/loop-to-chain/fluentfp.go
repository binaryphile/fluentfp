//go:build ignore

package main

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
)

type User struct {
	Name, Email, Role string
	Age               int
	Active            bool
}

// Method expressions for fluentfp
func (u User) IsActive() bool   { return u.Active }
func (u User) GetEmail() string { return u.Email }
func (u User) IsAdmin() bool    { return u.Role == "admin" }
func (u User) GetAge() int      { return u.Age }

// 1. Filter active users (3 lines)
func getActiveUsers(users []User) []User {
	return slice.From(users).KeepIf(User.IsActive)
}

// 2. Extract emails (3 lines)
func getEmails(users []User) []string {
	return slice.From(users).ToString(User.GetEmail)
}

// 3. Count admins (3 lines)
func countAdmins(users []User) int {
	return slice.From(users).KeepIf(User.IsAdmin).Len()
}

// 4. Average age (5 lines)
func averageAge(users []User) float64 {
	// sumInt adds two integers.
	sumInt := func(a, b int) int { return a + b }
	sum := slice.Fold(slice.From(users).ToInt(User.GetAge), 0, sumInt)
	return float64(sum) / float64(len(users))
}

// 5. Send notifications - STAYS AS LOOP (7 lines)
func sendNotifications(users []User) error {
	for _, u := range users {
		if err := sendEmail(u.Email); err != nil {
			return err
		}
	}
	return nil
}

// 6. Find by email - STAYS AS LOOP (7 lines)
func findByEmail(users []User, email string) *User {
	for i := range users {
		if users[i].Email == email {
			return &users[i]
		}
	}
	return nil
}

// 7. Group by role - STAYS AS LOOP (7 lines)
func groupByRole(users []User) map[string][]User {
	result := make(map[string][]User)
	for _, u := range users {
		result[u.Role] = append(result[u.Role], u)
	}
	return result
}

// 8. Process with retry - STAYS AS LOOP (7 lines)
func processWithRetry(users []User) {
	for _, u := range users {
		for attempt := 0; attempt < 3; attempt++ {
			if process(u) == nil {
				break
			}
		}
	}
}

// 9. Validate sequential IDs - STAYS AS LOOP (7 lines)
func validateSequentialIDs(users []User) bool {
	for i, u := range users {
		if u.Age != i+1 {
			return false
		}
	}
	return true
}

// 10. Read from channel - STAYS AS LOOP (7 lines)
func processFromChannel(ch <-chan User) {
	for u := range ch {
		fmt.Println(u.Name)
	}
}

// 11. Update in place - STAYS AS LOOP (7 lines)
func deactivateAll(users []User) {
	for i := range users {
		users[i].Active = false
	}
}

func sendEmail(string) error { return nil }
func process(User) error     { return nil }
