#!/bin/sh

echo "==> [Step 1/2] Dang nap du lieu tu MangaDex (Seeding)..."
# Chay seeder de dam bao database luon co du lieu moi
./mangahub-seeder

echo "==> [Step 2/2] Khoi dong Server MangaHub..."
# Chay server chinh
./mangahub-server
