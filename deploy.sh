set -e
pm2 stop anileha
go build
cd frontend && npm i && npm run build && cd ..
pm2 start anileha