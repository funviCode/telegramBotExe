package main

import (
	"bytes"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/exec"
	"strconv"
)

const (
	exePath      = "C:\\Users\\dima\\GolandProjects\\PricePulse\\PricePulse.exe"
	workDir      = "C:\\Users\\dima\\GolandProjects\\PricePulse"
	cronSchedule = "06 01 * * *" //cron: "минуты часы * * *"
)

func main() {

	token, chatID := loadConfig()
	bot := initBot(token)
	c := setupCron(bot, chatID, cronSchedule)

	// Запускаем cron планировщик
	c.Start()
	defer c.Stop()

	// В качестве примера обрабатываем входящие команды (например, запуск по запросу, в самом боте)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	go handleUpdates(bot, updates)
	select {}
}

// Загрузка данных из .env
func loadConfig() (string, int64) {
	if err := godotenv.Load("C:\\Users\\dima\\GolandProjects\\telegramBotExe\\.env"); err != nil {
		log.Fatal("ошибка в загрузке .env файла")
	}
	// Получаем переменные
	botToken := os.Getenv("TELEGRAM_TOKEN")
	botChatID := os.Getenv("TELEGRAM_CHAT_ID")

	// Проверка переменных
	if botToken == "" || botChatID == "" {
		log.Fatal("TELEGRAM_TOKEN or TELEGRAM_CHAT_ID не установлены")
	}

	// Конвертация ChatID в int64
	chatID, err := strconv.ParseInt(botChatID, 10, 64)
	if err != nil {
		log.Fatalf("некорректный ChatID: %v", err)
	}
	return botToken, chatID
}

// Инициализация бота
func initBot(token string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}
	log.Printf("Авторизован бот %s", bot.Self.UserName)
	return bot
}

// Создаем новый cron планировщик, расписание
func setupCron(bot *tgbotapi.BotAPI, chatID int64, schedule string) *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc(schedule, func() {
		log.Println("Запуск exe файла по расписанию...")
		if err := runExe(); err != nil {
			log.Printf("Ошибка при запуске exe: %v", err)
			sendMessage(bot, chatID, "Ошибка при запуске программы: "+err.Error())
		} else {
			sendMessage(bot, chatID, "Программа успешно запущена по расписанию!")
		}
	})
	if err != nil {
		log.Fatalf("Ошибка добавления задачи в cron: %v", err)
	}
	return c
}

func handleUpdates(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				sendMessage(bot, update.Message.Chat.ID, "Бот запущен. Программа будет запускаться по расписанию.")
			case "run":
				// Запускаем exe сразу по команде
				if err := runExe(); err != nil {
					sendMessage(bot, update.Message.Chat.ID, "Ошибка запуска программы: "+err.Error())
				} else {
					sendMessage(bot, update.Message.Chat.ID, "Программа успешно запущена!")
				}
			default:
				sendMessage(bot, update.Message.Chat.ID, "Неизвестная команда")
			}
		}
	}
}

// runExe запускает exe-файл и возвращает ошибку, если запуск не удался
func runExe() error {
	cmd := exec.Command(exePath)
	cmd.Dir = workDir // Рабочая директория

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // Перенаправляем стандартный вывод
	cmd.Stderr = &stderr // Перенаправляем ошибки

	err := cmd.Start()
	if err != nil {
		log.Printf("Ошибка запуска: %v", err)
		return err
	}

	// Можно дождаться завершения команды, если нужно сразу увидеть вывод
	if err = cmd.Wait(); err != nil {
		log.Printf("Команда завершилась с ошибкой: %v", err)
	}
	log.Printf("Вывод программы: %s", stdout.String())
	log.Printf("Ошибки программы: %s", stderr.String())
	return nil
}

// sendMessage отправляет сообщение через Telegram бота
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
