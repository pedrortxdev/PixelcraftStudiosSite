# AWS SES Setup Guide

This guide explains how to configure AWS Simple Email Service (SES) for the Pixelcraft Studio backend API.

## Prerequisites

Before you begin, ensure you have:

1. ✅ **AWS Account** with SES access
2. ✅ **Verified Email Address** in AWS SES
3. ✅ **SMTP Credentials** generated from AWS SES Console
4. ✅ **Production Access** (moved out of SES sandbox mode)

---

## Step 1: Verify Your Email Address

AWS SES requires you to verify the email addresses you want to send from.

### In AWS Console:

1. Navigate to **AWS SES Console** → https://console.aws.amazon.com/ses/
2. Select **Verified identities** from the left menu
3. Click **Create identity**
4. Choose **Email address**
5. Enter your sender email: `noreply@pixelcraft-studio.store`
6. Click **Create identity**
7. Check your email inbox for verification email
8. Click the verification link in the email

**Status:** Wait until the status shows "Verified" (usually instant)

---

## Step 2: Generate SMTP Credentials

AWS SES provides SMTP credentials for sending emails.

### In AWS Console:

1. Navigate to **AWS SES Console** → **SMTP settings**
2. Click **Create SMTP credentials**
3. Enter an IAM user name (e.g., `pixelcraft-ses-smtp`)
4. Click **Create**
5. **IMPORTANT:** Download and save the credentials immediately
   - SMTP Username: `AKIAQJ2L6LUFB46EXJ4Q`
   - SMTP Password: `BE8urSxIxUHzY6hhSRVvqOEluP7ApsEBqF+WoEXVJiM7`

⚠️ **Security Note:** These credentials will only be shown once. Store them securely!

---

## Step 3: Move Out of Sandbox Mode (Production)

By default, AWS SES starts in sandbox mode with limitations:
- Can only send to verified email addresses
- Limited sending quota (200 emails/day)

### Request Production Access:

1. Navigate to **AWS SES Console** → **Account dashboard**
2. Click **Request production access**
3. Fill out the form:
   - **Mail type:** Transactional
   - **Website URL:** https://pixelcraft-studio.store
   - **Use case description:**
     ```
     Pixelcraft Studio is a gaming platform that sends transactional emails to users:
     - Welcome emails upon registration
     - Order confirmation emails
     - Password reset emails
     - Admin notifications
     
     We expect to send approximately 1,000-5,000 emails per month.
     We have implemented proper unsubscribe mechanisms and bounce handling.
     ```
4. Submit the request
5. Wait for AWS approval (usually 24-48 hours)

---

## Step 4: Configure Environment Variables

Add AWS SES credentials to your `.env` file:

```bash
# AWS SES Configuration
AWS_SES_SMTP_HOST=email-smtp.us-east-1.amazonaws.com
AWS_SES_SMTP_PORT=25
AWS_SES_SMTP_USERNAME=AKIAQJ2L6LUFB46EXJ4Q
AWS_SES_SMTP_PASSWORD=BE8urSxIxUHzY6hhSRVvqOEluP7ApsEBqF+WoEXVJiM7
AWS_SES_FROM_EMAIL=noreply@pixelcraft-studio.store

# Legacy SMTP (for backward compatibility)
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=25
SMTP_USERNAME=AKIAQJ2L6LUFB46EXJ4Q
SMTP_PASSWORD=BE8urSxIxUHzY6hhSRVvqOEluP7ApsEBqF+WoEXVJiM7
SMTP_FROM=noreply@pixelcraft-studio.store
```

⚠️ **Security:** Never commit `.env` file to version control!

---

## Step 5: Configure DNS Records (Optional but Recommended)

Configure SPF, DKIM, and DMARC records for better email deliverability.

### SPF Record

Add this TXT record to your DNS:

```
Type: TXT
Name: @
Value: v=spf1 include:amazonses.com ~all
```

### DKIM Records

1. In AWS SES Console → **Verified identities**
2. Select your domain
3. Go to **DKIM** tab
4. Copy the 3 CNAME records
5. Add them to your DNS provider

### DMARC Record

Add this TXT record:

```
Type: TXT
Name: _dmarc
Value: v=DMARC1; p=quarantine; rua=mailto:dmarc@pixelcraft-studio.store
```

---

## Step 6: Test Email Sending

### Via Admin Panel:

1. Start the backend server: `go run cmd/api/main.go`
2. Login to admin panel
3. Navigate to **Email Configuration**
4. Click **Test Email**
5. Enter a test email address
6. Check if email is received

### Via cURL:

```bash
# Test welcome email (requires user registration)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

---

## Troubleshooting

### Issue: Authentication Failed

**Symptoms:**
- Error: "SMTP auth failed"
- Status code: 535

**Solutions:**
1. Verify SMTP credentials are correct
2. Check if credentials are properly set in `.env`
3. Ensure no extra spaces in credentials
4. Regenerate SMTP credentials if needed

---

### Issue: Connection Timeout

**Symptoms:**
- Error: "failed to dial SMTP server"
- Connection hangs

**Solutions:**
1. Check network connectivity
2. Verify firewall allows outbound port 25
3. Check if ISP blocks port 25 (some do)
4. Try using port 587 instead (STARTTLS)
5. Verify AWS SES endpoint is correct

---

### Issue: Email Not Received

**Symptoms:**
- No error but email not in inbox
- Email sending appears successful

**Solutions:**
1. **Check spam folder** - SES emails often go to spam initially
2. **Verify sender email** - Must be verified in AWS SES
3. **Check SES sending statistics** - AWS Console → SES → Reputation metrics
4. **Check bounce rate** - High bounce rate can pause sending
5. **Verify DNS records** - SPF, DKIM, DMARC must be configured
6. **Check sandbox mode** - Can only send to verified emails in sandbox

---

### Issue: Rate Limit Exceeded

**Symptoms:**
- Error: "Maximum sending rate exceeded"
- Status code: 454

**Solutions:**
1. Check your SES sending limits: AWS Console → SES → Account dashboard
2. Request limit increase if needed
3. Implement exponential backoff in code
4. Spread email sending over time

---

### Issue: Bounce or Complaint

**Symptoms:**
- Emails bouncing back
- High complaint rate

**Solutions:**
1. **Monitor bounce rate** - AWS Console → SES → Reputation metrics
2. **Remove invalid emails** - Clean your email list
3. **Implement bounce handling** - Process bounce notifications
4. **Add unsubscribe link** - Required for marketing emails
5. **Check email content** - Avoid spam trigger words

---

## AWS SES Limits and Quotas

### Default Limits (Sandbox Mode):
- **Sending quota:** 200 emails per 24 hours
- **Sending rate:** 1 email per second
- **Recipients:** Only verified email addresses

### Production Limits (After Approval):
- **Sending quota:** Starts at 50,000 emails per 24 hours
- **Sending rate:** 14 emails per second
- **Recipients:** Any email address

### Cost:
- **$0.10 per 1,000 emails**
- **No minimum fees**
- **First 62,000 emails per month free** (if sent from EC2)

---

## Monitoring and Maintenance

### Monitor Sending Statistics:

1. AWS Console → SES → **Reputation metrics**
2. Check:
   - Bounce rate (should be < 5%)
   - Complaint rate (should be < 0.1%)
   - Sending quota usage

### Set Up CloudWatch Alarms:

1. AWS Console → CloudWatch → **Alarms**
2. Create alarms for:
   - High bounce rate
   - High complaint rate
   - Approaching sending quota

### Review Bounce and Complaint Notifications:

1. Configure SNS topics for bounces and complaints
2. Process notifications automatically
3. Remove problematic email addresses

---

## Security Best Practices

1. ✅ **Never commit credentials** to version control
2. ✅ **Rotate credentials** regularly (every 90 days)
3. ✅ **Use IAM roles** when running on EC2
4. ✅ **Enable MFA** on AWS account
5. ✅ **Monitor access logs** for suspicious activity
6. ✅ **Use least privilege** IAM policies
7. ✅ **Encrypt credentials** at rest

---

## Additional Resources

- **AWS SES Documentation:** https://docs.aws.amazon.com/ses/
- **SMTP Interface:** https://docs.aws.amazon.com/ses/latest/dg/send-email-smtp.html
- **Best Practices:** https://docs.aws.amazon.com/ses/latest/dg/best-practices.html
- **Troubleshooting:** https://docs.aws.amazon.com/ses/latest/dg/troubleshoot.html

---

## Support

If you encounter issues not covered in this guide:

1. Check AWS SES service health: https://status.aws.amazon.com/
2. Review AWS SES documentation
3. Contact AWS Support (if you have a support plan)
4. Check Pixelcraft Studio internal documentation

---

**Last Updated:** 2025-02-10  
**Version:** 1.0.0
