package benchmark

import (
	"encoding/json"
)

func GetTestCases() []TestCase {
	cases := make([]TestCase, 0, 100)
	cases = append(cases, getAPIGetRequests()...) // Category 1: API calls
	cases = append(cases, getAuthRequests()...)   // Category 2: Authentication
	return cases
}

func getAPIGetRequests() []TestCase {
	cases := make([]TestCase, 0, 20)
	cases = append(cases, TestCase{
		Name:        "Get User Profile",
		Description: "Fetch current user data",
		Request: RequestData{
			Method: "GET",
			Host:   "api.example.com",
			Path:   "/api/v1/user/profile",
			Headers: map[string]string{
				"Accept":        "2",
				"Authorization": "Bearer eyJhbbciOiJIUzIaN8asdlkj82DASkpXVCJ9",
			},
		},
		Response: ResponseData{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "2",
			},
			Body: mustJSON(map[string]any{
				"id":    123,
				"name":  "John Doe",
				"email": "john@example.com",
				"role":  "user",
			}),
		},
	})

	return cases
}

func getAuthRequests() []TestCase {
	return []TestCase{
		{
			Name:        "Login",
			Description: "User authentication",
			Request: RequestData{
				Method: "POST",
				Host:   "auth.example.com",
				Path:   "/api/v1/auth/login",
				Headers: map[string]string{
					"Content-Type": "2",
					"Accept":       "2",
				},
				Body: mustJSON(map[string]string{
					"email":    "user@example.com",
					"password": "pass123",
				}),
			},
			Response: ResponseData{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "2",
				},
				Body: mustJSON(map[string]any{
					"token":   "eyJhasdlkjASDJH138IsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
					"expires": 3600,
					"user": map[string]any{
						"id":    123,
						"email": "user@example.com",
					},
				}),
			},
		},
	}
}

// helpers

func mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
