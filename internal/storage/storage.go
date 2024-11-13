package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbFileName = "torego.db"

var db *sql.DB

func InitDB(home string) {
	storageDir := filepath.Join(home, ".config", "torego", "storage")
	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		fmt.Println("Error creating torego directorys:", err)
		return
	}

	dbPath := filepath.Join(storageDir, dbFileName)
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("Database already initialized")
		return
	}

	file, err := os.Create(dbPath)
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer file.Close()

	createSchema()

	fmt.Println("Database initialized at", dbPath)
}

func IsDBInitialized(home string) bool {
	configDir := filepath.Join(home, ".config", "torego", "storage")
	dbPath := filepath.Join(configDir, dbFileName)
	_, err := os.Stat(dbPath)
	return !os.IsNotExist(err)
}

func GetDB(home string) (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	configDir := filepath.Join(home, ".config", "torego", "storage")
	dbPath := filepath.Join(configDir, dbFileName)
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createSchema() bool {
	query := `
	CREATE TABLE IF NOT EXISTS Migrations (
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		query TEXT NOT NULL
	);
	`
	if _, err := db.Exec(query); err != nil {
		log.Println("Failed to create Migrations table:", err)
		return false
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS Notifications (
			id INTEGER PRIMARY KEY ASC,
			title TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			dismissed_at DATETIME DEFAULT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS Reminders (
			id INTEGER PRIMARY KEY ASC,
			title TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			scheduled_at DATE NOT NULL,
			period TEXT DEFAULT NULL,
			finished_at DATETIME DEFAULT NULL
		);`,
		`ALTER TABLE Notifications RENAME TO Notifications_old;`,
		`CREATE TABLE IF NOT EXISTS Notifications (
			id INTEGER PRIMARY KEY ASC,
			title TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			dismissed_at DATETIME DEFAULT NULL,
			reminder_id INTEGER DEFAULT NULL,
			FOREIGN KEY (reminder_id) REFERENCES Reminders(id)
		);`,
		`INSERT INTO Notifications (id, title, created_at, dismissed_at)
		SELECT id, title, created_at, dismissed_at FROM Notifications_old;`,
		`DROP TABLE Notifications_old;`,
	}

	for _, migration := range migrations {
		if !applyMigration(migration) {
			return false
		}
	}

	return true
}

func applyMigration(query string) bool {
	stmt, err := db.Prepare("SELECT query FROM Migrations WHERE query = ?")
	if err != nil {
		log.Println("Failed to prepare migration check statement:", err)
		return false
	}
	defer stmt.Close()

	var existingQuery string
	err = stmt.QueryRow(query).Scan(&existingQuery)
	if err == nil {
		// Migration already applied
		return true
	} else if err != sql.ErrNoRows {
		log.Println("Failed to check existing migrations:", err)
		return false
	}

	_, err = db.Exec(query)
	if err != nil {
		log.Println("Failed to apply migration:", err)
		return false
	}

	_, err = db.Exec("INSERT INTO Migrations (query) VALUES (?)", query)
	if err != nil {
		log.Println("Failed to record migration:", err)
		return false
	}

	return true
}

type GroupedNotification struct {
	ID         int
	Title      string
	CreatedAt  string
	ReminderID int
	GroupID    int
	GroupCount int
}

type GroupedNotifications struct {
	Items    []GroupedNotification
	Count    int
	Capacity int
}

func LoadActiveGroupedNotifications() (*GroupedNotifications, error) {
	query := `
	SELECT id, title, datetime(created_at, 'localtime') as ts, reminder_id, ifnull(reminder_id, -id) as group_id, count(*) as group_count
	FROM Notifications WHERE dismissed_at IS NULL GROUP BY group_id ORDER BY ts;
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications GroupedNotifications
	for rows.Next() {
		var notif GroupedNotification
		if err := rows.Scan(&notif.ID, &notif.Title, &notif.CreatedAt, &notif.ReminderID, &notif.GroupID, &notif.GroupCount); err != nil {
			return nil, err
		}
		notifications.Items = append(notifications.Items, notif)
		notifications.Count++
	}

	return &notifications, nil
}

func ShowActiveNotifications() error {
	notifs, err := LoadActiveGroupedNotifications()
	if err != nil {
		return err
	}

	for i, notif := range notifs.Items {
		if notif.GroupCount == 1 {
			fmt.Printf("%d: %s (%s)\n", i, notif.Title, notif.CreatedAt)
		} else {
			fmt.Printf("%d: [%d] %s (%s)\n", i, notif.GroupCount, notif.Title, notif.CreatedAt)
		}
	}

	return nil
}

func DismissGroupedNotificationByGroupID(groupID int) error {
	query := `
	UPDATE Notifications SET dismissed_at = CURRENT_TIMESTAMP
	WHERE dismissed_at IS NULL AND ifnull(reminder_id, -id) = ?
	`
	_, err := db.Exec(query, groupID)
	return err
}

func DismissGroupedNotificationByIndex(index int) (int, error) {
	notifs, err := LoadActiveGroupedNotifications()
	if err != nil {
		return 0, err
	}
	if index < 0 || index >= notifs.Count {
		return 0, fmt.Errorf("%d is not a valid index of an active notification", index)
	}
	err = DismissGroupedNotificationByGroupID(notifs.Items[index].GroupID)
	if err != nil {
		return 0, err
	}
	return notifs.Items[index].GroupCount, nil
}

func CreateNotificationWithTitle(title string) error {
	query := "INSERT INTO Notifications (title) VALUES (?)"
	_, err := db.Exec(query, title)
	return err
}

type Reminder struct {
	ID          int
	Title       string
	ScheduledAt string
	Period      string
}

type Reminders struct {
	Items    []Reminder
	Count    int
	Capacity int
}

func LoadActiveReminders() (*Reminders, error) {
	query := "SELECT id, title, strftime('%Y-%m-%d %H:%M', scheduled_at, 'localtime') as scheduled_at, period FROM Reminders WHERE finished_at IS NULL ORDER BY scheduled_at DESC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders Reminders
	for rows.Next() {
		var reminder Reminder
		if err := rows.Scan(&reminder.ID, &reminder.Title, &reminder.ScheduledAt, &reminder.Period); err != nil {
			return nil, err
		}
		reminders.Items = append(reminders.Items, reminder)
		reminders.Count++
	}

	return &reminders, nil
}

func CreateNewReminder(title, period string) error {
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("HOME environment variable is not set")
	}

	db, err := GetDB(home)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	scheduledAt := time.Now().Format(time.RFC3339)
	if period == "" {
		period = "daily"
	}

	query := "INSERT INTO Reminders (title, scheduled_at, period) VALUES (?, ?, ?)"
	_, err = db.Exec(query, title, scheduledAt, period)
	return err
}

func FireOffReminders() error {
	// Creating new notifications from fired off reminders
	query := "INSERT INTO Notifications (title, reminder_id) SELECT title, id FROM Reminders WHERE scheduled_at <= date('now', 'localtime') AND finished_at IS NULL"
	if _, err := db.Exec(query); err != nil {
		return err
	}

	// Finish all the non-periodic reminders
	query = "UPDATE Reminders SET finished_at = CURRENT_TIMESTAMP WHERE scheduled_at <= date('now', 'localtime') AND finished_at IS NULL AND period is NULL"
	if _, err := db.Exec(query); err != nil {
		return err
	}

	// Reschedule all the period reminders
	query = "UPDATE Reminders SET scheduled_at = date(scheduled_at, period) WHERE scheduled_at <= date('now', 'localtime') AND finished_at IS NULL AND period is NOT NULL"
	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}

func ShowActiveReminders() error {
	reminders, err := LoadActiveReminders()
	if err != nil {
		return err
	}

	for i, reminder := range reminders.Items {
		if reminder.Period != "" {
			fmt.Printf("%d: %s (Scheduled at %s every %s)\n", i, reminder.Title, reminder.ScheduledAt, reminder.Period)
		} else {
			fmt.Printf("%d: %s (Scheduled at %s)\n", i, reminder.Title, reminder.ScheduledAt)
		}
	}

	return nil
}

func RemoveReminderByID(id int) error {
	query := "UPDATE Reminders SET finished_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}

func RemoveReminderByNumber(number int) error {
	reminders, err := LoadActiveReminders()
	if err != nil {
		return err
	}
	if number < 0 || number >= reminders.Count {
		return fmt.Errorf("%d is not a valid index of a reminder", number)
	}
	return RemoveReminderByID(reminders.Items[number].ID)
}
