package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// $ date '+%Y%m%d%H%M'

// not really a migration
var migrationInitSchema = gormigrate.Migration{
	ID: "202002192100",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			Artist{},
			Track{},
			User{},
			Setting{},
			Play{},
			Album{},
			Playlist{},
			PlayQueue{},
		).
			Error
	},
}

// not really a migration
var migrationCreateInitUser = gormigrate.Migration{
	ID: "202002192019",
	Migrate: func(tx *gorm.DB) error {
		const (
			initUsername = "admin"
			initPassword = "admin"
		)
		err := tx.
			Where("name=?", initUsername).
			First(&User{}).
			Error
		if !gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return tx.Create(&User{
			Name:     initUsername,
			Password: initPassword,
			IsAdmin:  true,
		}).
			Error
	},
}

var migrationMergePlaylist = gormigrate.Migration{
	ID: "202002192222",
	Migrate: func(tx *gorm.DB) error {
		if !tx.HasTable("playlist_items") {
			return nil
		}
		return tx.Exec(`
			UPDATE playlists
			SET items=( SELECT group_concat(track_id) FROM (
				SELECT track_id
				FROM playlist_items
				WHERE playlist_items.playlist_id=playlists.id
				ORDER BY created_at
			) );
			DROP TABLE playlist_items;`,
		).
			Error
	},
}

var migrationCreateTranscode = gormigrate.Migration{
	ID: "202003111222",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			TranscodePreference{},
		).
			Error
	},
}

var migrationAddGenre = gormigrate.Migration{
	ID: "202003121330",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			Genre{},
			Album{},
			Track{},
		).
			Error
	},
}

var migrationUpdateTranscodePrefIDX = gormigrate.Migration{
	ID: "202003241509",
	Migrate: func(tx *gorm.DB) error {
		var hasIDX int
		tx.
			Select("1").
			Table("sqlite_master").
			Where("type = ?", "index").
			Where("name = ?", "idx_user_id_client").
			Count(&hasIDX)
		if hasIDX == 1 {
			// index already exists
			return nil
		}
		step := tx.Exec(`
			ALTER TABLE transcode_preferences RENAME TO transcode_preferences_orig;
		`)
		if err := step.Error; err != nil {
			return fmt.Errorf("step rename: %w", err)
		}
		step = tx.AutoMigrate(
			TranscodePreference{},
		)
		if err := step.Error; err != nil {
			return fmt.Errorf("step create: %w", err)
		}
		step = tx.Exec(`
			INSERT INTO transcode_preferences (user_id, client, profile)
				SELECT user_id, client, profile
				FROM transcode_preferences_orig;
			DROP TABLE transcode_preferences_orig;
		`)
		if err := step.Error; err != nil {
			return fmt.Errorf("step copy: %w", err)
		}
		return nil
	},
}

var migrationAddAlbumIDX = gormigrate.Migration{
	ID: "202004302006",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			// should add new index for older installations
			Album{},
		).
			Error
	},
}

var migrationDeleteTrackArtistCascade = gormigrate.Migration{
	ID: "202005031842",
	Migrate: func(tx *gorm.DB) error {
		// based on https://www.sqlite.org/lang_altertable.html
		// step 1	get the current schema version
		// step 2	enable writable_schema
		// step 3	replace the sql
		// step 4	increment the current schema version
		// step 5	disable writable_schema
		var schemaVersion int
		err := tx.
			Raw("PRAGMA schema_version").
			Row().
			Scan(&schemaVersion)
		if err != nil {
			return fmt.Errorf("step get schema version: %w", err)
		}
		step := tx.Exec(`PRAGMA writable_schema=ON`)
		if err := step.Error; err != nil {
			return fmt.Errorf("step set writable_schema: %w", err)
		}
		step = tx.Exec(`
			UPDATE sqlite_master
			SET sql=REPLACE(sql, 'artists(id) ON DELETE CASCADE', 'artists(id)')
			WHERE type='table' AND name='tracks'
		`)
		if err := step.Error; err != nil {
			return fmt.Errorf("step set replace sql: %w", err)
		}
		schemaVersion++
		// tried to use `tx.Exec("PRAGMA schema_version=?", schemaVersion)` here
		// but it was giving me a syntax error, so here is some good ole printf sql
		step = tx.Raw(fmt.Sprintf("PRAGMA schema_version=%d", schemaVersion))
		if err := step.Error; err != nil {
			return fmt.Errorf("step inc schema version: %w", err)
		}
		step = tx.Exec(`PRAGMA writable_schema=OFF`)
		if err := step.Error; err != nil {
			return fmt.Errorf("step unset writable_schema: %w", err)
		}
		return nil
	},
}
