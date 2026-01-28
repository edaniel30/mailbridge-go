package gmail

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder_Basic(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "From filter",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().From("boss@company.com")
			},
			expected: "from:boss@company.com",
		},
		{
			name: "To filter",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().To("employee@company.com")
			},
			expected: "to:employee@company.com",
		},
		{
			name: "Subject without spaces",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Subject("invoice")
			},
			expected: "subject:invoice",
		},
		{
			name: "Subject with spaces",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Subject("monthly invoice")
			},
			expected: "subject:\"monthly invoice\"",
		},
		{
			name: "HasWords without spaces",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasWords("urgent")
			},
			expected: "urgent",
		},
		{
			name: "HasWords with spaces",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasWords("urgent meeting")
			},
			expected: "\"urgent meeting\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryBuilder_Status(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "IsUnread",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().IsUnread()
			},
			expected: "is:unread",
		},
		{
			name: "IsRead",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().IsRead()
			},
			expected: "is:read",
		},
		{
			name: "IsStarred",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().IsStarred()
			},
			expected: "is:starred",
		},
		{
			name: "IsImportant",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().IsImportant()
			},
			expected: "is:important",
		},
		{
			name: "IsSnoozed",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().IsSnoozed()
			},
			expected: "is:snoozed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryBuilder_Attachments(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "HasAttachment",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasAttachment()
			},
			expected: "has:attachment",
		},
		{
			name: "Filename",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Filename("report.pdf")
			},
			expected: "filename:report.pdf",
		},
		{
			name: "HasDrive",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasDrive()
			},
			expected: "has:drive",
		},
		{
			name: "HasDocument",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasDocument()
			},
			expected: "has:document",
		},
		{
			name: "HasSpreadsheet",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasSpreadsheet()
			},
			expected: "has:spreadsheet",
		},
		{
			name: "HasPresentation",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasPresentation()
			},
			expected: "has:presentation",
		},
		{
			name: "HasYouTube",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().HasYouTube()
			},
			expected: "has:youtube",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryBuilder_Dates(t *testing.T) {
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "After date",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().After(testDate)
			},
			expected: "after:2024/01/01",
		},
		{
			name: "Before date",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Before(testDate)
			},
			expected: "before:2024/01/01",
		},
		{
			name: "OlderThan 2 days",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().OlderThan("2d")
			},
			expected: "older_than:2d",
		},
		{
			name: "NewerThan 4 months",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().NewerThan("4m")
			},
			expected: "newer_than:4m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryBuilder_Size(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "LargerThan 5MB",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().LargerThan(5 * 1024 * 1024)
			},
			expected: "larger:5242880",
		},
		{
			name: "LargerThan using helper",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().LargerThan(MegaBytes(5))
			},
			expected: "larger:5242880",
		},
		{
			name: "SmallerThan 1MB",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().SmallerThan(1024 * 1024)
			},
			expected: "smaller:1048576",
		},
		{
			name: "SmallerThan using helper",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().SmallerThan(KiloBytes(500))
			},
			expected: "smaller:512000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryBuilder_Location(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "InInbox",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().InInbox()
			},
			expected: "in:inbox",
		},
		{
			name: "InTrash",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().InTrash()
			},
			expected: "in:trash",
		},
		{
			name: "InSpam",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().InSpam()
			},
			expected: "in:spam",
		},
		{
			name: "InSent",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().InSent()
			},
			expected: "in:sent",
		},
		{
			name: "InDrafts",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().InDrafts()
			},
			expected: "in:drafts",
		},
		{
			name: "Label",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Label("work")
			},
			expected: "label:work",
		},
		{
			name: "Category",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Category("primary")
			},
			expected: "category:primary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryBuilder_Complex(t *testing.T) {
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "Unread from boss with attachment",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().
					IsUnread().
					From("boss@company.com").
					HasAttachment()
			},
			expected: "is:unread from:boss@company.com has:attachment",
		},
		{
			name: "Invoice after date larger than 5MB",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().
					Subject("invoice").
					After(testDate).
					LargerThan(MegaBytes(5))
			},
			expected: "subject:invoice after:2024/01/01 larger:5242880",
		},
		{
			name: "Starred or important in inbox",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().
					IsStarred().
					OR().
					IsImportant().
					InInbox()
			},
			expected: "is:starred OR is:important in:inbox",
		},
		{
			name: "NOT from specific sender",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().
					NOT().
					From("spam@example.com").
					IsUnread()
			},
			expected: "NOT from:spam@example.com is:unread",
		},
		{
			name: "Complex query with multiple conditions",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().
					From("client@company.com").
					Subject("proposal").
					HasAttachment().
					After(testDate).
					IsUnread().
					InInbox()
			},
			expected: "from:client@company.com subject:proposal has:attachment after:2024/01/01 is:unread in:inbox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryBuilder_Reset(t *testing.T) {
	qb := NewQueryBuilder().From("test@example.com").IsUnread()
	assert.Equal(t, "from:test@example.com is:unread", qb.Build())

	qb.Reset().Subject("invoice")
	assert.Equal(t, "subject:invoice", qb.Build())
}

func TestQueryBuilder_Raw(t *testing.T) {
	qb := NewQueryBuilder().
		From("boss@company.com").
		Raw("custom:query").
		IsUnread()

	assert.Equal(t, "from:boss@company.com custom:query is:unread", qb.Build())
}

func TestQueryBuilder_Recipients(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name: "Cc",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Cc("manager@company.com")
			},
			expected: "cc:manager@company.com",
		},
		{
			name: "Bcc",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().Bcc("hidden@company.com")
			},
			expected: "bcc:hidden@company.com",
		},
		{
			name: "DeliveredTo",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().DeliveredTo("group@company.com")
			},
			expected: "deliveredto:group@company.com",
		},
		{
			name: "List",
			builder: func() *QueryBuilder {
				return NewQueryBuilder().List("dev-team@company.com")
			},
			expected: "list:dev-team@company.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSizeHelpers(t *testing.T) {
	tests := []struct {
		name     string
		helper   func() int64
		expected int64
	}{
		{
			name:     "KiloBytes",
			helper:   func() int64 { return KiloBytes(10) },
			expected: 10 * 1024,
		},
		{
			name:     "MegaBytes",
			helper:   func() int64 { return MegaBytes(5) },
			expected: 5 * 1024 * 1024,
		},
		{
			name:     "GigaBytes",
			helper:   func() int64 { return GigaBytes(2) },
			expected: 2 * 1024 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.helper()
			assert.Equal(t, tt.expected, result)
		})
	}
}
