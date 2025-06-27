build-darwin-arm64:
	@echo "Building for darwin/arm64..."
	@start_time=$$(date +%s)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 GO111MODULE=on go build -a -o warp-server cmd/app/*
	echo  "\033[1;32mBuild succeeded in $$(($$(date +%s) - start_time)) seconds\033[0m"
	chmod +x warp-server

build:
	go build -o warp-server cmd/app/*
	chmod +x warp-server

plist: agent # Создаем .plist-файл
	./make_plist.sh
	mv application.ru.server.warp.daemon.plist $(HOME)/Library/LaunchAgents/application.ru.server.warp.daemon.plist

agent: build-darwin-arm64 # Помещаем исполняемый файл в каталог из перечня поиска PATH TODO: можно переменить
	sudo cp warp-server /usr/local/bin/warp-server

# ---
# Управление демоном
#

start-daemon: # Запуск вручную
	launchctl start application.ru.server.warp.daemon

stop-daemon: # Остановка
	launchctl stop application.ru.server.warp.daemon

unload-daemon:  # Перезагрузка демона
	-launchctl remove application.ru.server.warp.daemon
	-launchctl bootout gui/$(shell id -u)/application.ru.server.warp.daemon

load-daemon: plist # Загружаем демона
	launchctl bootstrap gui/$(shell id -u) $(HOME)/Library/LaunchAgents/application.ru.server.warp.daemon.plist

reload-daemon: unload-daemon load-daemon
	@echo "\033[1;32m✔ Daemon has been reloaded!\033[0m"

config: # Сформировать конфигурацию по умолчанию HOME/.warp-server.yaml TODO: можно переменить
	./make_config.sh

plist-lint: # Проверяем синтаксис
	plutil -lint $(HOME)/Library/LaunchAgents/application.ru.server.warp.daemon.plist

check-daemon: # Проверить статус
	-launchctl list | grep warp

describe-daemon: # Или более подробно
	-launchctl print gui/$UID/application.ru.server.warp.daemon

logs: # Просмотреть логи
	cat /tmp/warp-server.{out,err}

clear-logs: # Почистить логи
	rm -rf /tmp/warp-server.{out,err}
