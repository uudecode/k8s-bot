
# 🤖 k8s-bot
**k8s-bot** — это инструмент для мониторинга Kubernetes-кластера и статуса приложений в ArgoCD с уведомлениями в Telegram.
## 🚀 Возможности
 * **Мониторинг нод:** Следит за состоянием нод кластера и нагрузкой.
 * **ArgoCD Watcher:** Отслеживает статус синхронизации (Sync Status) и здоровья (Health Status) приложений.
 * **Уведомления:** Мгновенная отправка отчетов в Telegram.
## 🛠 Конфигурация
Бот использует config.toml для настроек, а чувствительные данные (токены) забирает из переменных окружения.
### Переменные окружения
Перед запуском установите следующие переменные:
 * TELEGRAM_BOT_TOKEN — токен вашего бота.
 * TELEGRAM_CHAT_ID — ID чата для уведомлений.
 * K8S_API_URL — адрес API вашего кластера.
 * K8S_SERVICE_TOKEN — токен сервисного аккаунта для доступа к API.
### config.toml
Пример файла конфигурации:
```toml
[monitoring]
check_interval = "60s"
nodes_enabled = true
argocd_enabled = true

[namespaces]
include = ["default", "production"]

```
## 📋 Архитектура
## 🚀 Запуск
 1. **Клонирование:**
   ```bash
   git clone https://github.com/your-username/k8s-bot.git
   cd k8s-bot
   
   ```
 2. **Сборка:**
   ```bash
   go build -o k8s-bot main.go
   
   ```
 3. **Запуск:**
   ```bash
   # Экспорт переменных и запуск
   export TELEGRAM_BOT_TOKEN="ваш_токен"
   export K8S_SERVICE_TOKEN="ваш_токен_са"
   
   ./k8s-bot --config=config.toml
   
   ```
*Сделано с любовью к Kubernetes и Go.*

