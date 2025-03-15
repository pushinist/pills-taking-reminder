Привет, это мое решение предложенного задания.

Используемые сторонние библиотеки: chi, cleanenv, pq, godotenv, validator

Для самого задания использовал Go, Postgres, Docker

Соответственно, для запуска нужно склонировать репозиторий, отредактировать .env и config.yaml в случае необходимости (для работы достаточно раскомменитровать .example, так запустится :) ) в корне проекта запустить docker compose и после этого сервер готов принимать запросы:

`git clone https://github.com/pushinist/pills-taking-reminder.git && cd pills-taking-reminder && mv .env.example .env && mv config/local.yaml.example config/local.yaml && docker compose up`
