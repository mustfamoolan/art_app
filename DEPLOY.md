# دليل نشر وتشغيل السيرفر وتوجيه الدومين (Continuous Line String Art Backend)

يحتوي هذا الملف على الخطوات الدقيقة لربط الدومين الخاص بك `marcetime.com` وتشغيل الخلفية باستخدام Docker على سيرفر الـ VPS ذو العنوان `72.60.180.203`.

---

## 🛠️ الخطوة 1: ضبط سجلات الدومين (DNS Records)

اذهب إلى لوحة تحكم نطاقك (على موقع تسجيل الدومين مثل GoDaddy أو Namecheap أو Cloudflare) وقم بإضافة/تعديل السجلات التالية:

1. **سجل النطاق الرئيسي (Root Domain):**
   - **النوع (Type):** `A`
   - **الاسم (Host/Name):** `@`
   - **القيمة (Value/Points to):** `72.60.180.203`
   
2. **سجل النطاق الفرعي (WWW Domain):**
   - **النوع (Type):** `A` (أو `CNAME` يشير إلى `marcetime.com`)
   - **الاسم (Host/Name):** `www`
   - **القيمة (Value/Points to):** `72.60.180.203`

---

## 💻 الخطوة 2: تهيئة سيرفر الـ VPS وتشغيل الحاويات (Docker)

اتصل بالسيرفر الخاص بك عبر الـ SSH:
```bash
ssh root@72.60.180.203
```

### 1. جلب كود المشروع من GitHub:
قم بسحب المشروع داخل السيرفر (إذا لم تكن قد فعلت ذلك بعد):
```bash
git clone https://github.com/mustfamoolan/art_app.git project
cd project
```

### 2. إعداد ملف البيئة `.env`:
أنشئ ملف `.env` واملأه بالمتغيرات اللازمة بناءً على `.env.example`:
```bash
cp .env.example .env
nano .env
```
*(قم بتعديل قيم كلمات مرور قاعدة البيانات `DB_PASSWORD` وكلمة سر التطبيق `APP_KEY` بما يناسبك ثم احفظ الملف بالضغط على `Ctrl+O` ثم `Enter` ثم الخروج بـ `Ctrl+X`).*

### 3. بناء وتشغيل الحاويات:
شغّل مشروعك في الخلفية باستخدام Docker Compose:
```bash
docker compose up --build -d
```
تأكد من عمل الحاويات بنجاح عبر الأمر:
```bash
docker compose ps
```

---

## 🌐 الخطوة 3: إعداد Nginx كـ Reverse Proxy

سنقوم بتثبيت Nginx ليقوم باستقبال طلبات النطاق `marcetime.com` وتمريرها للتطبيق الداخلي العامل على منفذ `8080`.

### 1. تثبيت Nginx:
```bash
sudo apt update
sudo apt install nginx -y
```

### 2. إضافة إعدادات التوجيه:
أنشئ ملف تهيئة جديد:
```bash
sudo nano /etc/nginx/sites-available/marcetime.com
```

ألصق الإعدادات التالية داخل الملف:
```nginx
server {
    listen 80;
    server_name marcetime.com www.marcetime.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
*(احفظ الملف واخرج).*

### 3. تفعيل الموقع وإعادة تشغيل Nginx:
```bash
sudo ln -s /etc/nginx/sites-available/marcetime.com /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## 🔒 الخطوة 4: تفعيل شهادة الأمان (SSL / HTTPS)

لتشغيل الاتصال الآمن (مهم جداً للاتصال من تطبيق الموبايل):

### 1. تثبيت أداة Certbot الخاصة بـ Let's Encrypt:
```bash
sudo apt install certbot python3-certbot-nginx -y
```

### 2. إصدار شهادة الأمان:
نفّذ الأمر التالي ودع الأداة تقوم بتهيئة النطاق تلقائياً:
```bash
sudo certbot --nginx -d marcetime.com -d www.marcetime.com
```
*عندما يسألك عما إذا كنت ترغب في تحويل حركة المرور غير الآمنة (HTTP) تلقائياً إلى الآمنة (HTTPS)، اختر الموافقة والتفعيل (Redirect).*

مبروك! سيرفرك الآن مهيأ ومحمي بالكامل ويعمل عبر الرابط: `https://marcetime.com`
