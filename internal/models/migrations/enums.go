package migrations

import (
	"gorm.io/gorm"
)

// CreateEnumTypes creates all required ENUM types in the database
func CreateEnumTypes(db *gorm.DB) error {
	enumDefinitions := []struct {
		Name   string   // ENUM型の名前
		Values []string // ENUM型に含める値
	}{
		{"material_status", []string{"draft", "published", "archived"}},
		{"importance_level", []string{"low", "medium", "high"}},
		{"word_level", []string{"beginner", "intermediate", "advanced"}},
		{"progress_status", []string{"not_started", "in_progress", "completed"}},
		{"sender_type", []string{"user", "system", "bot"}},
	}

	for _, enum := range enumDefinitions {
		if err := db.Exec(createEnumSQL(enum.Name, enum.Values)).Error; err != nil {
			return err
		}
	}

	return nil
}

// createEnumSQL generates the SQL to create an ENUM type if it doesn't exist
func createEnumSQL(name string, values []string) string {
	quotedValues := "'" + join(values, "','") + "'"
	return `
		DO $$ BEGIN
			CREATE TYPE ` + name + ` AS ENUM (` + quotedValues + `);
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`
}

// join is a helper function to join strings with a delimiter
func join(values []string, delimiter string) string {
	result := ""
	for i, v := range values {
		if i > 0 {
			result += delimiter
		}
		result += v
	}
	return result
}
