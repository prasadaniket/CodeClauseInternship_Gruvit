package com.gruvit.auth.dto;

import jakarta.validation.constraints.NotBlank;

public class TwoFactorRequest {
    @NotBlank
    private String token;
    
    @NotBlank
    private String code;
    
    public String getToken() { return token; }
    public void setToken(String token) { this.token = token; }
    public String getCode() { return code; }
    public void setCode(String code) { this.code = code; }
}
