curl -X POST http://localhost:8000/login -d "username=testuser&password=testpass" -i

curl -X GET http://localhost:8000/notes \
     -H "Cookie: token=your_jwt_token_here"


curl -X POST http://localhost:8000/notes \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -H "Cookie: token=your_jwt_token_here" \
     -d "text=Your note text here"
