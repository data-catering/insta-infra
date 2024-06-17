# Customization Guide

## Overview

Insta-Infra provides a flexible and customizable environment for deploying and managing various services on your local laptop. This guide will help you tailor the setup to meet your specific needs.

## Customizing Services

### Adding Custom Data

You can add custom data, such as startup SQL scripts, to your services. This is useful for initializing databases with default values or configurations.

1. **Locate the Data Directory:** Navigate to the `data` directory.
2. **Add Your Script:** Place your custom script in the appropriate directory.

### Environment Variables

Set environment variables to customize the behavior of your services.

1. **Locate Environment Files:** Find the `.env` file or environment section in the `docker-compose.yaml`.
2. **Add or Modify Variables:** Add or modify the variables as needed.
