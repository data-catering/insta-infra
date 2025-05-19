#!/bin/bash

# Wait for CVAT server to be ready
echo "Waiting for CVAT server to be ready..."
until curl -s http://cvat-server:8080/api/v1/server/health; do
    echo "CVAT server not ready yet. Waiting..."
    sleep 5
done

echo "CVAT server is ready. Creating default superuser..."

# Create a temporary Python script to create the superuser
cat > /tmp/create_superuser.py << EOF
import os
import sys
import django
from django.contrib.auth.models import User

# Create superuser if it doesn't exist
username = os.getenv('CVAT_USER', 'admin')
email = os.getenv('CVAT_EMAIL', 'admin@example.com')
password = os.getenv('CVAT_PASSWORD', 'admin')

try:
    if not User.objects.filter(username=username).exists():
        User.objects.create_superuser(username=username, email=email, password=password)
        print(f"Created superuser {username}")
    else:
        # Update existing user's password
        user = User.objects.get(username=username)
        user.set_password(password)
        user.save()
        print(f"Updated password for existing user {username}")
except Exception as e:
    print(f"Error creating/updating user: {e}")
    sys.exit(1)
EOF

# Execute the Python script using manage.py
cd /home/django && python3 manage.py shell < /tmp/create_superuser.py

echo "Default superuser created successfully!" 