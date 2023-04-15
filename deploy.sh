set -e
pm2 stop anileha
git pull
go build
cd frontend && npm i && npm run build && cd ..
pm2 start anileha