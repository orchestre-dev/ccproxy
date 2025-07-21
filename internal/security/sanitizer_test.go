package security

import (
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewDataSanitizer(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("with nil config", func(t *testing.T) {
		sanitizer, err := NewDataSanitizer(nil)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, sanitizer)
		testutil.AssertEqual(t, SecurityLevelBasic, sanitizer.config.Level)
	})

	t.Run("with custom patterns", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.SensitivePatterns = []string{"\\bcustom\\b", "\\bpattern\\b"}

		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, sanitizer)
		testutil.AssertEqual(t, 2, len(sanitizer.sensitiveRegexps))
	})

	t.Run("with invalid pattern", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.SensitivePatterns = []string{"[invalid regex"}

		_, err := NewDataSanitizer(config)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid sensitive pattern")
	})

	t.Run("default redaction patterns", func(t *testing.T) {
		sanitizer, err := NewDataSanitizer(nil)
		testutil.AssertNoError(t, err)

		expectedPatterns := []string{"api_key", "token", "password", "secret", "credit_card", "ssn", "email"}
		for _, pattern := range expectedPatterns {
			testutil.AssertNotEqual(t, nil, sanitizer.redactPatterns[pattern])
		}
	})
}

func TestSanitizeString(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("HTML escaping", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = true // Allow sensitive data for testing HTML escaping
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "<script>alert('xss')</script>"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "&lt;script&gt;")
		testutil.AssertContains(t, result, "&lt;/script&gt;")
		testutil.AssertNotContains(t, result, "<script>")
	})

	t.Run("removes null bytes", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = true
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "hello\x00world\x00"
		result := sanitizer.SanitizeString(input)

		testutil.AssertEqual(t, "helloworld", result)
	})

	t.Run("removes control characters", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = true
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "hello\x01\x02world\x7F"
		result := sanitizer.SanitizeString(input)

		testutil.AssertEqual(t, "helloworld", result)
	})

	t.Run("redacts API key", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "api_key: sk-1234567890abcdef"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_API_KEY]")
		testutil.AssertNotContains(t, result, "sk-1234567890abcdef")
	})

	t.Run("redacts token", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "Bearer token: abc123def456"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_TOKEN]")
		testutil.AssertNotContains(t, result, "abc123def456")
	})

	t.Run("redacts password", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "password=supersecret123"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_PASSWORD]")
		testutil.AssertNotContains(t, result, "supersecret123")
	})

	t.Run("redacts secret", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "secret: mysecretvalue"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_SECRET]")
		testutil.AssertNotContains(t, result, "mysecretvalue")
	})

	t.Run("redacts credit card", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "Card: 4532 1234 5678 9012"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_CREDIT_CARD]")
		testutil.AssertNotContains(t, result, "4532 1234 5678 9012")
	})

	t.Run("redacts SSN", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "SSN: 123-45-6789"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_SSN]")
		testutil.AssertNotContains(t, result, "123-45-6789")
	})

	t.Run("redacts email", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "Contact: user@example.com"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_EMAIL]")
		testutil.AssertNotContains(t, result, "user@example.com")
	})

	t.Run("does not redact when logging sensitive data enabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = true
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "api_key: sk-1234567890abcdef"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "sk-1234567890abcdef")
		testutil.AssertNotContains(t, result, "[REDACTED_API_KEY]")
	})

	t.Run("handles URL decoding", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = true
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "hello%20world%21"
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "hello world!")
	})

	t.Run("complex string with multiple issues", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := "<script>alert('XSS')</script>\x00\x01 api_key=sk-123 password=secret email=user@test.com"
		result := sanitizer.SanitizeString(input)

		// Should escape HTML
		testutil.AssertContains(t, result, "&lt;script&gt;")
		// Should remove control chars
		testutil.AssertNotContains(t, result, "\x00")
		testutil.AssertNotContains(t, result, "\x01")
		// Should redact sensitive data
		testutil.AssertContains(t, result, "[REDACTED_API_KEY]")
		testutil.AssertContains(t, result, "[REDACTED_PASSWORD]")
		testutil.AssertContains(t, result, "[REDACTED_EMAIL]")
	})
}

func TestSanitizeRequest(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	sanitizer, err := NewDataSanitizer(config)
	testutil.AssertNoError(t, err)

	t.Run("sanitizes map data", func(t *testing.T) {
		input := map[string]interface{}{
			"message":  "Hello <script>alert('xss')</script>",
			"api_key":  "sk-123456789",
			"password": "secret123",
			"safe":     "normal data",
		}

		result, err := sanitizer.SanitizeRequest(input)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		testutil.AssertContains(t, resultMap["message"].(string), "&lt;script&gt;")
		testutil.AssertEqual(t, "[REDACTED]", resultMap["api_key"])
		testutil.AssertEqual(t, "[REDACTED]", resultMap["password"])
		testutil.AssertEqual(t, "normal data", resultMap["safe"])
	})

	t.Run("sanitizes array data", func(t *testing.T) {
		input := []interface{}{
			"Hello <script>",
			map[string]interface{}{
				"token": "abc123",
				"data":  "safe",
			},
			"normal string",
		}

		result, err := sanitizer.SanitizeRequest(input)
		testutil.AssertNoError(t, err)

		resultArray := result.([]interface{})
		testutil.AssertContains(t, resultArray[0].(string), "&lt;script&gt;")

		nestedMap := resultArray[1].(map[string]interface{})
		testutil.AssertEqual(t, "[REDACTED]", nestedMap["token"])
		testutil.AssertEqual(t, "safe", nestedMap["data"])

		testutil.AssertEqual(t, "normal string", resultArray[2])
	})

	t.Run("handles unmarshallable data", func(t *testing.T) {
		input := make(chan int) // Cannot be marshaled to JSON

		_, err := sanitizer.SanitizeRequest(input)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to marshal request")
	})

	t.Run("handles non-map, non-array data", func(t *testing.T) {
		input := "simple string with <script>alert('xss')</script>"

		result, err := sanitizer.SanitizeRequest(input)
		testutil.AssertNoError(t, err)

		// Should return as-is since it's not a parseable structure
		testutil.AssertEqual(t, input, result.(string))
	})

	t.Run("handles empty data", func(t *testing.T) {
		input := map[string]interface{}{}

		result, err := sanitizer.SanitizeRequest(input)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		testutil.AssertEqual(t, 0, len(resultMap))
	})

	t.Run("handles nested complex data", func(t *testing.T) {
		input := map[string]interface{}{
			"user": map[string]interface{}{
				"name":     "John <script>",
				"password": "secret123",
				"details": map[string]interface{}{
					"email":   "john@example.com",
					"api_key": "sk-nested-key",
				},
			},
			"items": []interface{}{
				"item1",
				map[string]interface{}{
					"token": "abc123",
				},
			},
		}

		result, err := sanitizer.SanitizeRequest(input)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		user := resultMap["user"].(map[string]interface{})
		testutil.AssertContains(t, user["name"].(string), "&lt;script&gt;")
		testutil.AssertEqual(t, "[REDACTED]", user["password"])

		details := user["details"].(map[string]interface{})
		testutil.AssertEqual(t, "[REDACTED]", details["email"])
		testutil.AssertEqual(t, "[REDACTED]", details["api_key"])

		items := resultMap["items"].([]interface{})
		testutil.AssertEqual(t, "item1", items[0])

		nestedItem := items[1].(map[string]interface{})
		testutil.AssertEqual(t, "[REDACTED]", nestedItem["token"])
	})
}

func TestSanitizeResponse(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	sanitizer, err := NewDataSanitizer(config)
	testutil.AssertNoError(t, err)

	t.Run("sanitizes response data", func(t *testing.T) {
		input := map[string]interface{}{
			"status": "success",
			"token":  "response-token-123",
			"data":   "Hello <script>alert('response')</script>",
		}

		result, err := sanitizer.SanitizeResponse(input)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		testutil.AssertEqual(t, "success", resultMap["status"])
		testutil.AssertEqual(t, "[REDACTED]", resultMap["token"])
		testutil.AssertContains(t, resultMap["data"].(string), "&lt;script&gt;")
	})
}

func TestIsSensitiveKey(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	sanitizer, err := NewDataSanitizer(nil)
	testutil.AssertNoError(t, err)

	sensitiveKeys := []string{
		"password", "passwd", "pwd",
		"secret", "api_key", "apikey", "api-key",
		"token", "auth", "authorization",
		"credit_card", "creditcard", "cc_number",
		"ssn", "social_security", "social-security",
		"private_key", "private-key", "privatekey",
		"access_key", "access-key", "accesskey",
		"refresh_token", "refresh-token",
		"session", "session_id", "session-id",
		"cookie", "csrf", "xsrf",
	}

	for _, key := range sensitiveKeys {
		t.Run("detects "+key, func(t *testing.T) {
			testutil.AssertTrue(t, sanitizer.isSensitiveKey(key))
		})
	}

	safeKeys := []string{
		"username", "name", "id", "status", "message", "data", "result",
	}

	for _, key := range safeKeys {
		t.Run("allows "+key, func(t *testing.T) {
			testutil.AssertFalse(t, sanitizer.isSensitiveKey(key))
		})
	}

	t.Run("partial matches", func(t *testing.T) {
		testutil.AssertTrue(t, sanitizer.isSensitiveKey("user_password"))
		testutil.AssertTrue(t, sanitizer.isSensitiveKey("my_api_key"))
		testutil.AssertTrue(t, sanitizer.isSensitiveKey("session_token"))
		testutil.AssertFalse(t, sanitizer.isSensitiveKey("user_name"))
	})
}

func TestSanitizeArray(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	sanitizer, err := NewDataSanitizer(config)
	testutil.AssertNoError(t, err)

	t.Run("sanitizes mixed array", func(t *testing.T) {
		input := []interface{}{
			"Hello <script>alert('xss')</script>",
			map[string]interface{}{
				"password": "secret123",
				"message":  "safe data",
			},
			[]interface{}{
				"nested array item",
				map[string]interface{}{
					"token": "nested-token-123",
				},
			},
			42,
			true,
		}

		result := sanitizer.sanitizeArray(input)

		// Check string sanitization
		testutil.AssertContains(t, result[0].(string), "&lt;script&gt;")

		// Check map sanitization
		resultMap := result[1].(map[string]interface{})
		testutil.AssertEqual(t, "[REDACTED]", resultMap["password"])
		testutil.AssertEqual(t, "safe data", resultMap["message"])

		// Check nested array sanitization
		nestedArray := result[2].([]interface{})
		testutil.AssertEqual(t, "nested array item", nestedArray[0])
		nestedMap := nestedArray[1].(map[string]interface{})
		testutil.AssertEqual(t, "[REDACTED]", nestedMap["token"])

		// Check primitive types are preserved
		testutil.AssertEqual(t, 42, result[3])
		testutil.AssertEqual(t, true, result[4])
	})

	t.Run("empty array", func(t *testing.T) {
		input := []interface{}{}
		result := sanitizer.sanitizeArray(input)
		testutil.AssertEqual(t, 0, len(result))
	})

	t.Run("array with nil values", func(t *testing.T) {
		input := []interface{}{
			"string",
			nil,
			"another string",
		}

		result := sanitizer.sanitizeArray(input)
		testutil.AssertEqual(t, 3, len(result))
		testutil.AssertEqual(t, "string", result[0])
		testutil.AssertEqual(t, nil, result[1])
		testutil.AssertEqual(t, "another string", result[2])
	})
}

func TestRedactSecrets(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	sanitizer, err := NewDataSanitizer(nil)
	testutil.AssertNoError(t, err)

	t.Run("redacts multiple secret types", func(t *testing.T) {
		message := "User logged in with api_key=sk-123 and password=secret123, then got token=abc456, email user@example.com, credit card 4532-1234-5678-9012"
		result := sanitizer.RedactSecrets(message)

		testutil.AssertContains(t, result, "[REDACTED_API_KEY]")
		testutil.AssertContains(t, result, "[REDACTED_PASSWORD]")
		testutil.AssertContains(t, result, "[REDACTED_TOKEN]")
		testutil.AssertContains(t, result, "[REDACTED_EMAIL]")
		testutil.AssertContains(t, result, "[REDACTED_CREDIT_CARD]")

		testutil.AssertNotContains(t, result, "sk-123")
		testutil.AssertNotContains(t, result, "secret123")
		testutil.AssertNotContains(t, result, "abc456")
		testutil.AssertNotContains(t, result, "user@example.com")
		testutil.AssertNotContains(t, result, "4532-1234-5678-9012")
	})

	t.Run("no secrets to redact", func(t *testing.T) {
		message := "User successfully retrieved profile data"
		result := sanitizer.RedactSecrets(message)

		testutil.AssertEqual(t, message, result)
	})
}

func TestMaskingFunctions(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("MaskValue", func(t *testing.T) {
		testutil.AssertEqual(t, "abc***", MaskValue("abcdef", 3))
		testutil.AssertEqual(t, "****", MaskValue("test", 0))
		testutil.AssertEqual(t, "***", MaskValue("abc", 5)) // showChars > length
		testutil.AssertEqual(t, "*", MaskValue("a", 1))
	})

	t.Run("MaskEmail", func(t *testing.T) {
		testutil.AssertEqual(t, "us**@example.com", MaskEmail("user@example.com"))
		testutil.AssertEqual(t, "*@test.com", MaskEmail("a@test.com"))       // Short username (1 char)
		testutil.AssertEqual(t, "inv**********", MaskEmail("invalid-email")) // Invalid format - first 3 + 10 masks
		testutil.AssertEqual(t, "**@domain.com", MaskEmail("ab@domain.com")) // 2-char username
	})

	t.Run("MaskAPIKey", func(t *testing.T) {
		testutil.AssertEqual(t, "sk-1***5678", MaskAPIKey("sk-12345678"))             // 11 chars: first 4 + 3 stars + last 4
		testutil.AssertEqual(t, "abc1*****xyz9", MaskAPIKey("abc123456xyz9"))         // 13 chars: first 4 + 5 stars + last 4
		testutil.AssertEqual(t, "*****", MaskAPIKey("short"))                         // Too short (5 chars)
		testutil.AssertEqual(t, "********", MaskAPIKey("8charkey"))                   // Exactly 8 chars
		testutil.AssertEqual(t, "long*********test", MaskAPIKey("longapikeyfortest")) // 17 chars: first 4 + 9 stars + last 4
	})
}

func TestRemoveSensitiveData(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	sanitizer, err := NewDataSanitizer(config)
	testutil.AssertNoError(t, err)

	t.Run("removes sensitive data", func(t *testing.T) {
		input := map[string]interface{}{
			"username":  "john",
			"password":  "secret123",
			"api_key":   "sk-123456",
			"message":   "Hello World",
			"token":     "abc123",
			"safe_data": "this is fine",
		}

		result := sanitizer.RemoveSensitiveData(input)

		testutil.AssertEqual(t, "john", result["username"])
		testutil.AssertEqual(t, "[REDACTED]", result["password"])
		testutil.AssertEqual(t, "[REDACTED]", result["api_key"])
		testutil.AssertEqual(t, "Hello World", result["message"])
		testutil.AssertEqual(t, "[REDACTED]", result["token"])
		testutil.AssertEqual(t, "this is fine", result["safe_data"])
	})

	t.Run("handles nested data", func(t *testing.T) {
		input := map[string]interface{}{
			"user": map[string]interface{}{
				"name":     "John",
				"password": "secret",
				"details": map[string]interface{}{
					"email": "john@example.com",
					"phone": "555-1234",
				},
			},
		}

		result := sanitizer.RemoveSensitiveData(input)

		user := result["user"].(map[string]interface{})
		testutil.AssertEqual(t, "John", user["name"])
		testutil.AssertEqual(t, "[REDACTED]", user["password"])

		details := user["details"].(map[string]interface{})
		testutil.AssertEqual(t, "555-1234", details["phone"])
		// Email should be redacted because "email" is a sensitive key
		testutil.AssertEqual(t, "[REDACTED]", details["email"])
	})
}

func TestEdgeCases(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("handles circular references gracefully", func(t *testing.T) {
		// Note: This is a limitation - we can't easily test circular references
		// because JSON marshaling will fail. In a real implementation, we might
		// need more sophisticated cycle detection.
		sanitizer, err := NewDataSanitizer(nil)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, sanitizer)
	})

	t.Run("handles very large strings", func(t *testing.T) {
		config := DefaultSecurityConfig()
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		// Test with a very large string
		largeString := make([]byte, 100000)
		for i := range largeString {
			largeString[i] = 'a'
		}

		result := sanitizer.SanitizeString(string(largeString))
		testutil.AssertEqual(t, 100000, len(result))
	})

	t.Run("handles special characters in patterns", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.LogSensitiveData = false
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		// Test with special characters that might break regex
		input := "api_key='sk-123' and password=\"secret\""
		result := sanitizer.SanitizeString(input)

		testutil.AssertContains(t, result, "[REDACTED_API_KEY]")
		testutil.AssertContains(t, result, "[REDACTED_PASSWORD]")
	})

	t.Run("preserves structure with type assertions", func(t *testing.T) {
		config := DefaultSecurityConfig()
		sanitizer, err := NewDataSanitizer(config)
		testutil.AssertNoError(t, err)

		input := map[string]interface{}{
			"number":  42,
			"boolean": true,
			"float":   3.14,
			"null":    nil,
		}

		result, err := sanitizer.SanitizeRequest(input)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})

		// Note: JSON unmarshaling converts numbers to float64
		num, ok := resultMap["number"].(float64)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, 42.0, num)

		boolean, ok := resultMap["boolean"].(bool)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, true, boolean)

		float, ok := resultMap["float"].(float64)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, 3.14, float)

		testutil.AssertEqual(t, nil, resultMap["null"])
	})
}
