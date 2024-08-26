
## Инструкция по использованию сервиса

1. Для регистрации необходимо выполнить следующий запрос:
```
curl -X POST http://localhost:8000/login -d "username=имя_пользователя&password=пароль" -i
```

2. Для авторизации необходимо выполнить следующий запрос:
```
curl -X POST http://localhost:8000/login -d "username=testuser&password=testpass" -i
```

3. Для авторизации необходимо выполнить следующий запрос:
```
curl -X GET http://localhost:8000/notes \
     -H "Cookie: token=ваш_jwt_токен"
```

4. Для авторизации необходимо выполнить следующий запрос:
```
curl -X POST http://localhost:8000/notes \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -H "Cookie: token=ваш_jwt_токен" \
     -d "text=текст_вашей_заметки"
```