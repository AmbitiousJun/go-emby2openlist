cd web/src

npm -i

npm run build

cd ..

rm -rf ./dist
mv src/build/client ./dist
