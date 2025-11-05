# 🚀 Quick Start Guide

## Prerequisites

- Python 3.9+ installed
- NHIT Backend API running on `http://localhost:8080`
- pip package manager

## 5-Minute Setup

### Step 1: Navigate to Test Directory
```bash
cd NHIT-Backend-API-Tests/python-pytest
```

### Step 2: Create Virtual Environment
```bash
python -m venv venv
```

### Step 3: Activate Virtual Environment

**Windows:**
```bash
venv\Scripts\activate
```

**Linux/Mac:**
```bash
source venv/bin/activate
```

### Step 4: Install Dependencies
```bash
pip install -r requirements.txt
```

### Step 5: Configure Environment
```bash
# Copy example env file
cp .env.example .env

# Edit .env with your settings (optional - defaults work for local)
```

### Step 6: Run Tests!
```bash
pytest -v
```

## 🎯 Quick Test Commands

### Run Smoke Tests (Fast)
```bash
pytest -m smoke -v
```

### Run All Tests with Report
```bash
pytest --html=reports/report.html --self-contained-html
```

### Run Specific Test File
```bash
pytest tests/test_users.py -v
```

## 📊 Expected Output

```
======================== test session starts ========================
collected 45 items

tests/test_users.py::TestUserManagement::test_create_user_success PASSED
tests/test_users.py::TestUserManagement::test_get_user_success PASSED
tests/test_auth.py::TestAuthentication::test_login_success PASSED
...

======================== 45 passed in 12.34s ========================
```

## 🎉 That's It!

Your API tests are now running. Check `reports/report.html` for detailed results.

## 🔧 Troubleshooting

### API Not Running?
```bash
# Start your API Gateway
cd path/to/NHIT-Backend/services/api-gateway
go run cmd/server/main.go
```

### Connection Refused?
- Check `API_BASE_URL` in `.env`
- Ensure API is running on port 8080
- Try: `curl http://localhost:8080/api/v1/users`

### Import Errors?
```bash
# Reinstall dependencies
pip install -r requirements.txt --force-reinstall
```

## 📚 Next Steps

- Read [README.md](README.md) for detailed documentation
- Explore test files in `tests/` directory
- Add your own tests
- Integrate with CI/CD

Happy Testing! 🎊
