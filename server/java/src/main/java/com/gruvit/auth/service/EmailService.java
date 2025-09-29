package com.gruvit.auth.service;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.mail.SimpleMailMessage;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.stereotype.Service;
import java.util.UUID;

@Service
public class EmailService {
    
    @Autowired
    private JavaMailSender mailSender;
    
    @Value("${app.email.from:noreply@gruvit.com}")
    private String fromEmail;
    
    @Value("${app.base.url:http://localhost:8080}")
    private String baseUrl;
    
    public void sendVerificationEmail(String toEmail, String username, String token) {
        try {
            SimpleMailMessage message = new SimpleMailMessage();
            message.setFrom(fromEmail);
            message.setTo(toEmail);
            message.setSubject("Verify your Gruvit account");
            
            String verificationUrl = baseUrl + "/auth/verify-email?token=" + token;
            String emailBody = String.format("""
                Hello %s,
                
                Welcome to Gruvit! Please verify your email address by clicking the link below:
                
                %s
                
                This link will expire in 24 hours.
                
                If you didn't create an account with Gruvit, please ignore this email.
                
                Best regards,
                The Gruvit Team
                """, username, verificationUrl);
            
            message.setText(emailBody);
            mailSender.send(message);
            
        } catch (Exception e) {
            // Log error but don't throw exception to avoid breaking user registration
            System.err.println("Failed to send verification email: " + e.getMessage());
        }
    }
    
    public void sendPasswordResetEmail(String toEmail, String username, String token) {
        try {
            SimpleMailMessage message = new SimpleMailMessage();
            message.setFrom(fromEmail);
            message.setTo(toEmail);
            message.setSubject("Reset your Gruvit password");
            
            String resetUrl = baseUrl + "/auth/reset-password?token=" + token;
            String emailBody = String.format("""
                Hello %s,
                
                You requested to reset your password. Click the link below to reset it:
                
                %s
                
                This link will expire in 1 hour.
                
                If you didn't request a password reset, please ignore this email.
                
                Best regards,
                The Gruvit Team
                """, username, resetUrl);
            
            message.setText(emailBody);
            mailSender.send(message);
            
        } catch (Exception e) {
            // Log error but don't throw exception
            System.err.println("Failed to send password reset email: " + e.getMessage());
        }
    }
    
    public String generateToken() {
        return UUID.randomUUID().toString().replace("-", "");
    }
}
