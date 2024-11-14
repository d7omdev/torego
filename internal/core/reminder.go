package core

import (
	"fmt"
	"time"

	"github.com/d7omdev/torego/internal/storage"
)

type Reminder struct {
	ID          int
	Title       string
	ScheduledAt string
	Period      string
}

func CreateReminder(title string, period string) error {
	return storage.CreateNewReminder(title, period)
}

func GetReminder(id int) (*storage.Reminder, error) {
	reminders, err := storage.LoadActiveReminders()
	if err != nil {
		return nil, err
	}
	for _, reminder := range reminders.Items {
		if reminder.ID == id {
			return &reminder, nil
		}
	}
	return nil, fmt.Errorf("no reminder found with id %d", id)
}

func GetAllReminders() (*storage.Reminders, error) {
	return storage.LoadActiveReminders()
}

func UpdateReminder(id int, title string, scheduledAt time.Time, period string) error {
	err := storage.RemoveReminderByID(id)
	if err != nil {
		return err
	}
	return storage.CreateNewReminder(title, period)
}

func DeleteReminder(id int) error {
	return storage.RemoveReminderByID(id)
}

func ListReminders() ([]storage.Reminder, error) {
	reminders, err := storage.LoadActiveReminders()
	if err != nil {
		return nil, err
	}
	return reminders.Items, nil
}

func CheckoutReminder() error {
	return storage.FireOffReminders()
}

func ShowActiveNotifications() error {
	reminders, err := ListReminders()
	if err != nil {
		return err
	}
	if len(reminders) == 0 {
		fmt.Println("No reminders found.")
		fmt.Println("Use 'torego remind <title> [period]' to set a reminder.")
		return nil
	}

	for _, reminder := range reminders {
		fmt.Printf("%d: %s\n", reminder.ID, reminder.Title)
	}

	return nil
}
