import psycopg2
from psycopg2 import OperationalError

try:
    conn = psycopg2.connect(
        dbname="vietnam_chain",
        user="vietnam_chain",
        password="vietnam_chain",
        host="localhost",
        port=5432
    )
    print("✅ Kết nối thành công tới PostgreSQL!")
    conn.close()
except OperationalError as e:
    print("❌ Lỗi kết nối:", e)
