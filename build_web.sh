cd web/src

npm ci
npm run build

cd ..

rm -rf ./dist
mv src/build/client ./dist

cd ..