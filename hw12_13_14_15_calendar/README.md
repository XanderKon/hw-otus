## Локальный запуск

Понадобится:

 - golang ~1.21;
 - PSQL ~14.

Используется файл **Makefile** как исходная точка для запуска приложения, серверов и других сопутствующих вещей.

Заполняем файл конфигурации `configs/config.toml` согласно потребностям.

| Параметр  	| Описание                          	| Пример                            	|
|-----------	|-----------------------------------	|-----------------------------------	|
| [logger]  	|                                   	|                                   	|
| level     	| Уровень логирования               	| DEBUG \| INFO \| WARNING \| ERROR 	|
| path      	| Путь к файлу для логирования      	| "./logs/calendar.log"             	|
| [storage] 	|                                   	|                                   	|
| driver    	| Драйвер для хранилища             	| memory \| postgres                	|
| [db]      	|                                   	|                                   	|
| host      	| Хост для драйвера postgres        	| "localhost"                       	|
| port      	| Порт для драйвера postgres        	| 5432                              	|
| name      	| Название БД для драйвера postgres 	| "otus-db"                         	|
| username  	| Логин для драйвера postgres       	| "postgres"                        	|
| password  	| Пароль для драйвера postgres      	| "postgres"                        	|
| [http]    	|                                   	|                                   	|
| host      	| Хост для HTTP сервера             	| "localhost"                       	|
| port      	| Порт для HTTP сервера             	| 8080                              	|
| [grpc]    	|                                   	|                                   	|
| host      	| Хост для GRPC сервера             	| "localhost"                       	|


Для запуска **ВНЕ** Docker выполняем:

 - `make run`

Результатом выполнения команды будут два сервера Http и Grpc, располженных на портах, согласно файлу конфигурации `configs/config.toml`

Для тестирования запросов можно использовать коллекцию Postman: [http-collection](docs/postman/OTUS.HTTP.postman_collection.json)

Для тестирования GRPC-сервера можно воспользоваться функционалом **ServerReflectionInfo** (https://github.com/grpc/grpc/blob/master/src/proto/grpc/reflection/v1alpha/reflection.proto), который позволит автоматически подгрузить доступные для вызова методы.