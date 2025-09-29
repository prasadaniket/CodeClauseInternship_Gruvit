package com.gruvit.auth.service;

import com.google.zxing.BarcodeFormat;
import com.google.zxing.WriterException;
import com.google.zxing.client.j2se.MatrixToImageWriter;
import com.google.zxing.common.BitMatrix;
import com.google.zxing.qrcode.QRCodeWriter;
import org.apache.commons.codec.binary.Base32;
import org.springframework.stereotype.Service;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.security.SecureRandom;
import java.time.Instant;

@Service
public class TwoFactorService {
    
    private static final String ISSUER = "Gruvit";
    private static final int SECRET_LENGTH = 20;
    
    public String generateSecret() {
        SecureRandom random = new SecureRandom();
        byte[] bytes = new byte[SECRET_LENGTH];
        random.nextBytes(bytes);
        Base32 base32 = new Base32();
        return base32.encodeToString(bytes);
    }
    
    public String generateQRCodeUrl(String username, String secret) {
        return String.format("otpauth://totp/%s:%s?secret=%s&issuer=%s",
                ISSUER, username, secret, ISSUER);
    }
    
    public byte[] generateQRCodeImage(String qrCodeUrl) throws WriterException, IOException {
        QRCodeWriter qrCodeWriter = new QRCodeWriter();
        BitMatrix bitMatrix = qrCodeWriter.encode(qrCodeUrl, BarcodeFormat.QR_CODE, 200, 200);
        
        ByteArrayOutputStream pngOutputStream = new ByteArrayOutputStream();
        MatrixToImageWriter.writeToStream(bitMatrix, "PNG", pngOutputStream);
        return pngOutputStream.toByteArray();
    }
    
    public boolean verifyCode(String secret, String code) {
        long timeIndex = Instant.now().getEpochSecond() / 30;
        return verifyCode(secret, code, timeIndex) ||
               verifyCode(secret, code, timeIndex - 1) ||
               verifyCode(secret, code, timeIndex + 1);
    }
    
    private boolean verifyCode(String secret, String code, long timeIndex) {
        try {
            Base32 base32 = new Base32();
            byte[] key = base32.decode(secret);
            byte[] time = new byte[8];
            for (int i = 7; i >= 0; i--) {
                time[i] = (byte) (timeIndex & 0xFF);
                timeIndex >>= 8;
            }
            
            byte[] hash = hmacSha1(key, time);
            int offset = hash[hash.length - 1] & 0x0F;
            int binary = ((hash[offset] & 0x7F) << 24) |
                        ((hash[offset + 1] & 0xFF) << 16) |
                        ((hash[offset + 2] & 0xFF) << 8) |
                        (hash[offset + 3] & 0xFF);
            
            int otp = binary % 1000000;
            String expectedCode = String.format("%06d", otp);
            
            return expectedCode.equals(code);
        } catch (Exception e) {
            return false;
        }
    }
    
    private byte[] hmacSha1(byte[] key, byte[] data) throws Exception {
        javax.crypto.Mac mac = javax.crypto.Mac.getInstance("HmacSHA1");
        javax.crypto.spec.SecretKeySpec secretKeySpec = new javax.crypto.spec.SecretKeySpec(key, "HmacSHA1");
        mac.init(secretKeySpec);
        return mac.doFinal(data);
    }
}
