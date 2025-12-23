# Використовуємо Python 3.9 або 3.10 (найстабільніші для ML бібліотек зараз)
FROM python:3.10-slim

# Оновлюємо список пакетів і встановлюємо системні залежності
# libgl1 і libglib2.0 потрібні для роботи OpenCV
# git потрібен, бо nomeroff-net може підтягувати щось з гіта
RUN apt-get update && apt-get install -y \
    libgl1 libglx-mesa0 \
    libglib2.0-0 \
    git \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копіюємо requirements і ставимо бібліотеки
COPY requirements.txt .

# Це може зайняти 5-10 хвилин, бо бібліотеки важкі
RUN pip install --no-cache-dir -r requirements.txt

RUN pip install "git+https://github.com/ria-com/upscaler.git" && \
    pip install "git+https://github.com/ria-com/craft-text-detector.git" && \
    pip install "git+https://github.com/lilohuang/PyTurboJPEG.git" && \
    pip install "git+https://github.com/ria-com/modelhub-client.git"

# Копіюємо код
COPY main.py .

# Створюємо папку для логів або кешу моделей, якщо потрібно
RUN mkdir -p /var/www/nomeroff-net

EXPOSE 8000

CMD ["python", "main.py"]