#!/bin/bash

# API测试脚本

BASE_URL="http://localhost:8080"
TEST_FILE="test.txt"

echo "🧪 开始API测试..."
echo ""

# 检查服务是否运行
echo "1. 检查服务健康状态..."
if curl -s "$BASE_URL/api/health" > /dev/null; then
    echo "✅ 服务运行正常"
else
    echo "❌ 服务未运行，请先启动服务"
    exit 1
fi

echo ""

# 测试文件上传
echo "2. 测试文件上传..."
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/files/upload" \
  -F "file=@$TEST_FILE")

echo "上传响应: $UPLOAD_RESPONSE"

# 提取文件ID
FILE_ID=$(echo $UPLOAD_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$FILE_ID" ]; then
    echo "✅ 文件上传成功，文件ID: $FILE_ID"
else
    echo "❌ 文件上传失败"
    exit 1
fi

echo ""

# 测试获取文件列表
echo "3. 测试获取文件列表..."
LIST_RESPONSE=$(curl -s "$BASE_URL/api/files")
echo "文件列表: $LIST_RESPONSE"
echo "✅ 文件列表获取成功"

echo ""

# 测试获取文件信息
echo "4. 测试获取文件信息..."
INFO_RESPONSE=$(curl -s "$BASE_URL/api/files/$FILE_ID/info")
echo "文件信息: $INFO_RESPONSE"
echo "✅ 文件信息获取成功"

echo ""

# 测试文件下载
echo "5. 测试文件下载..."
DOWNLOAD_FILE="downloaded_$TEST_FILE"
curl -s -o "$DOWNLOAD_FILE" "$BASE_URL/api/files/download/$FILE_ID"

if [ -f "$DOWNLOAD_FILE" ]; then
    echo "✅ 文件下载成功: $DOWNLOAD_FILE"
    echo "文件内容:"
    cat "$DOWNLOAD_FILE"
    echo ""
else
    echo "❌ 文件下载失败"
fi

echo ""

# 测试文件删除
echo "6. 测试文件删除..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/files/$FILE_ID")
echo "删除响应: $DELETE_RESPONSE"
echo "✅ 文件删除成功"

echo ""

# 验证文件已删除
echo "7. 验证文件已删除..."
INFO_RESPONSE_AFTER_DELETE=$(curl -s "$BASE_URL/api/files/$FILE_ID/info")
if echo "$INFO_RESPONSE_AFTER_DELETE" | grep -q "文件不存在"; then
    echo "✅ 文件删除验证成功"
else
    echo "❌ 文件删除验证失败"
fi

echo ""

# 清理测试文件
rm -f "$DOWNLOAD_FILE"

echo "🎉 所有测试完成！"
