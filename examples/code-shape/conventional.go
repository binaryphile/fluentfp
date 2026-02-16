package main

import "fmt"

type User struct {
	Name, Email, Role string
	Age               int
	Active            bool
}

// 1. Filter active users (7 lines)
func getActiveUsers(users []User) []User {
	var result []User
	for _, u := range users {
		if u.Active {
			result = append(result, u)
		}
	}
	return result
}

// 2. Extract emails (7 lines)
func getEmails(users []User) []string {
	var emails []string
	for _, u := range users {
		emails = append(emails, u.Email)
	}
	return emails
}

// 3. Count admins (7 lines)
func countAdmins(users []User) int {
	count := 0
	for _, u := range users {
		if u.Role == "admin" {
			count++
		}
	}
	return count
}

// 4. Average age (7 lines)
func averageAge(users []User) float64 {
	sum := 0
	for _, u := range users {
		sum += u.Age
	}
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
