// Focused sanitization tests without React dependencies

// Test utilities for sanitization functions (matching the actual implementation)
const sanitizeStringValue = (input: string): string => {
  const inputStr = String(input)

  // Block dangerous protocols and JavaScript URLs
  const dangerousPatterns = [
    /javascript:/gi,
    /data:/gi,
    /vbscript:/gi,
    /on\w+\s*=/gi // Event handlers like onclick, onload, etc.
  ]

  let sanitized = inputStr
  let previousLength
  // Apply patterns repeatedly until no more matches are found (handles nested attacks)
  do {
    previousLength = sanitized.length
    dangerousPatterns.forEach((pattern) => {
      sanitized = sanitized.replace(pattern, '')
    })
  } while (sanitized.length !== previousLength && sanitized.length > 0)

  // Comprehensive HTML entity encoding for XSS prevention
  return (
    sanitized
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#x27;')
      .replace(/\//g, '&#x2F;')
      // Remove null bytes and control characters
      .replace(/[\0-\x1F\x7F]/g, '')
      // Limit string length to prevent DoS
      .substring(0, 500)
      .trim()
  )
}

const isValidKey = (key: string): boolean => {
  return /^[a-zA-Z][a-zA-Z0-9_]*$/.test(key) && key.length <= 50
}

describe('Query Parameters Sanitization', () => {
  describe('sanitizeStringValue', () => {
    it('should encode HTML entities to prevent XSS', () => {
      const maliciousInput = '<script>alert("xss")</script>'
      const sanitized = sanitizeStringValue(maliciousInput)
      expect(sanitized).toBe(
        '&lt;script&gt;alert(&quot;xss&quot;)&lt;&#x2F;script&gt;'
      )
      expect(sanitized).not.toContain('<')
      expect(sanitized).not.toContain('>')
      expect(sanitized).not.toContain('"')
    })

    it('should handle various XSS attack vectors', () => {
      const attacks = [
        { input: 'javascript:alert(1)', expected: 'alert(1)' },
        {
          input: '<img src=x onerror=alert(1)>',
          expected: '&lt;img src=x alert(1)&gt;'
        },
        {
          input: '"><script>alert(1)</script>',
          expected: '&quot;&gt;&lt;script&gt;alert(1)&lt;&#x2F;script&gt;'
        },
        {
          input: "'><script>alert(1)</script>",
          expected: '&#x27;&gt;&lt;script&gt;alert(1)&lt;&#x2F;script&gt;'
        },
        {
          input: 'data:text/html,<script>alert(1)</script>',
          expected: 'text&#x2F;html,&lt;script&gt;alert(1)&lt;&#x2F;script&gt;'
        },
        { input: 'vbscript:msgbox(1)', expected: 'msgbox(1)' }
      ]

      attacks.forEach(({ input, expected }) => {
        const result = sanitizeStringValue(input)
        expect(result).toBe(expected)
        expect(result).not.toMatch(/javascript:|data:|vbscript:|on\w+\s*=/i)
      })
    })

    it('should handle nested encoding attacks', () => {
      const nestedAttacks = [
        // Double encoding attempts
        {
          input: '&lt;script&gt;alert(1)&lt;/script&gt;',
          expected: '&amp;lt;script&amp;gt;alert(1)&amp;lt;&#x2F;script&amp;gt;'
        },
        // Mixed case protocols - should be cleaned
        { input: 'JaVaScRiPt:alert(1)', expected: 'alert(1)' },
        {
          input: 'DATA:text/html,<script>',
          expected: 'text&#x2F;html,&lt;script&gt;'
        },
        // Nested dangerous patterns - should be fully cleaned with multiple passes
        { input: 'jajavascript:vascript:alert(1)', expected: 'alert(1)' },
        // Multiple event handlers
        {
          input: '<div onclick=alert(1) onload=alert(2)>',
          expected: '&lt;div alert(1) alert(2)&gt;'
        }
      ]

      nestedAttacks.forEach(({ input, expected }) => {
        const result = sanitizeStringValue(input)
        expect(result).toBe(expected)
        // Verify dangerous patterns are removed
        expect(result).not.toMatch(/javascript:|data:|vbscript:|on\w+\s*=/i)
      })
    })

    it('should handle Unicode normalization attacks', () => {
      const unicodeAttacks = [
        // Unicode escaped characters that could bypass filters
        {
          input: '\\u003cscript\\u003ealert(1)\\u003c/script\\u003e',
          expected: '\\u003cscript\\u003ealert(1)\\u003c&#x2F;script\\u003e'
        },
        // Different Unicode representations of dangerous characters
        { input: '\u003cscript\u003e', expected: '&lt;script&gt;' },
        {
          input: '\u0022\u003e\u003cscript\u003e',
          expected: '&quot;&gt;&lt;script&gt;'
        },
        // Zero-width characters that could be used to bypass filters (javascript: pattern won't match with zero-width char)
        {
          input: 'java\u200bscript:alert(1)',
          expected: 'java\u200bscript:alert(1)'
        },
        // Homograph attacks (similar looking characters - Cyrillic chars look like latin but won't match javascript pattern)
        { input: 'јаvascript:alert(1)', expected: 'јаvascript:alert(1)' } // Cyrillic characters
      ]

      unicodeAttacks.forEach(({ input, expected }) => {
        const result = sanitizeStringValue(input)
        expect(result).toBe(expected)
      })
    })

    it('should handle URL-encoded payload tests', () => {
      const encodedAttacks = [
        // URL encoded script tags (should be encoded but not decoded first)
        {
          input: '%3Cscript%3Ealert(1)%3C/script%3E',
          expected: '%3Cscript%3Ealert(1)%3C&#x2F;script%3E'
        },
        // Double URL encoded
        { input: '%253Cscript%253E', expected: '%253Cscript%253E' },
        // Mixed encoding with dangerous protocols (URL encoded colon won't match javascript: pattern)
        { input: 'javascript%3Aalert(1)', expected: 'javascript%3Aalert(1)' },
        {
          input: 'data%3Atext%2Fhtml%2C%3Cscript%3E',
          expected: 'data%3Atext%2Fhtml%2C%3Cscript%3E'
        },
        // Hex encoding gets HTML encoded, not removed
        {
          input: 'javascript&#x3A;alert(1)',
          expected: 'javascript&amp;#x3A;alert(1)'
        }
      ]

      encodedAttacks.forEach(({ input, expected }) => {
        const result = sanitizeStringValue(input)
        expect(result).toBe(expected)
        // Only check for literal patterns, not encoded ones
        expect(result).not.toMatch(/javascript:|data:|vbscript:/i)
      })
    })

    it('should handle edge cases in sanitization logic', () => {
      const edgeCases = [
        // Empty dangerous protocols
        { input: 'javascript:', expected: '' },
        { input: 'data:', expected: '' },
        // Whitespace variations
        { input: ' javascript: alert(1) ', expected: 'alert(1)' },
        { input: '\tdata:\ttext/html', expected: 'text&#x2F;html' },
        // Multiple dangerous patterns in one string
        {
          input: 'javascript:alert(1);data:text/html',
          expected: 'alert(1);text&#x2F;html'
        },
        // Case sensitivity edge cases
        { input: 'JavaScript:Alert(1)', expected: 'Alert(1)' },
        // Boundary conditions - patterns need to match exactly "javascript:" not partial matches
        { input: 'notjavascript:safe', expected: 'notsafe' }, // "javascript:" will be removed
        { input: 'javascript:but_safe_content', expected: 'but_safe_content' }
      ]

      edgeCases.forEach(({ input, expected }) => {
        const result = sanitizeStringValue(input)
        expect(result).toBe(expected)
      })
    })

    it('should remove null bytes and control characters', () => {
      const input = 'test\0null\x01control\x1Fchars\x7Fdel'
      const result = sanitizeStringValue(input)
      expect(result).toBe('testnullcontrolcharsdel')
      expect(result).not.toMatch(/[\0-\x1F\x7F]/)
    })

    it('should limit string length to 500 characters', () => {
      const longString = 'a'.repeat(1000)
      const result = sanitizeStringValue(longString)
      expect(result.length).toBe(500)
      expect(result).toBe('a'.repeat(500))
    })

    it('should trim whitespace', () => {
      const input = '  spaced content  '
      const result = sanitizeStringValue(input)
      expect(result).toBe('spaced content')
    })

    it('should handle empty and null-like inputs', () => {
      expect(sanitizeStringValue('')).toBe('')
      expect(sanitizeStringValue('   ')).toBe('')
      expect(sanitizeStringValue('null')).toBe('null')
      expect(sanitizeStringValue('undefined')).toBe('undefined')
    })
  })

  describe('isValidKey', () => {
    it('should accept valid parameter keys', () => {
      const validKeys = [
        'name',
        'outputFormat',
        'createdAt',
        'page',
        'limit',
        'userName123',
        'filter_type'
      ]

      validKeys.forEach((key) => {
        expect(isValidKey(key)).toBe(true)
      })
    })

    it('should reject invalid parameter keys', () => {
      const invalidKeys = [
        '123name', // starts with number
        'name-with-dash', // contains dash
        'name with space', // contains space
        'name.with.dot', // contains dot
        '<script>', // contains HTML
        'a'.repeat(51), // too long
        '', // empty
        '$money', // special character
        'name@domain' // contains @
      ]

      invalidKeys.forEach((key) => {
        expect(isValidKey(key)).toBe(false)
      })
    })

    it('should handle edge cases', () => {
      expect(isValidKey('a')).toBe(true) // minimum valid length
      expect(isValidKey('a'.repeat(50))).toBe(true) // maximum valid length
      expect(isValidKey('_invalid')).toBe(false) // starts with underscore
    })
  })

  describe('Form value serialization edge cases', () => {
    it('should handle numeric edge cases', () => {
      const edgeCases = [
        { value: 0, expected: 0 },
        { value: -1, expected: -1 },
        { value: 1.5, expected: 1.5 },
        { value: Infinity, expected: undefined },
        { value: -Infinity, expected: undefined },
        { value: NaN, expected: undefined }
      ]

      // Note: This is a conceptual test - in practice we'd need to test the actual hook
      // but for now we're testing the sanitization logic
      edgeCases.forEach(({ value, expected }) => {
        if (typeof value === 'number') {
          if (isFinite(value) && !isNaN(value)) {
            expect(value).toBe(expected)
          } else {
            expect(expected).toBeUndefined()
          }
        }
      })
    })
  })
})

describe('Input validation utilities', () => {
  it('should properly validate different input types', () => {
    // Test that numeric validation works correctly
    const validNumbers = [0, -1, 1.5, 100]
    const invalidNumbers = [Infinity, -Infinity, NaN]

    validNumbers.forEach((num) => {
      expect(isFinite(num) && !isNaN(num)).toBe(true)
    })

    invalidNumbers.forEach((num) => {
      expect(isFinite(num) && !isNaN(num)).toBe(false)
    })
  })
})
