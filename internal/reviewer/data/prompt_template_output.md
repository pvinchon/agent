## Output

Return valid JSON only.

Rules:

- Output must be valid JSON
- Do not include explanations
- Do not include markdown
- Do not include comments
- Do not include any text before or after the JSON
- The response must be directly parsable by a JSON parser

JSON Schema:
```json
[
  {
    "severity": "CRITICAL | HIGH | MEDIUM | LOW",
    "title": "string",
    "description": "string",
    "location": "string (format: filename:line)"
  }
]
```

Example JSON Output:
```json
[
  {
    "severity": "HIGH",
    "title": "Possible nil pointer dereference",
    "description": "The pointer is dereferenced without checking if it is nil. This can cause a runtime panic.",
    "location": "service/user.go:42"
  }
]
```

If the diff contains **no issues**, return:
```json
[]
```

Strict Requirement:
Your entire response must be a single JSON array.
The first character must be `[`.
The last character must be `]`.