package com.example.blog.config;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.JwtException;
import io.jsonwebtoken.Jwts;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import javax.crypto.SecretKey;
import javax.crypto.spec.SecretKeySpec;
import java.io.IOException;
import java.nio.charset.StandardCharsets;

/**
 * Verifies the same HS256 access tokens stakeholders issues and exposes the
 * caller's identity as request attributes, instead of trusting the
 * X-Username header the gateway sets (which itself never checks the JWT
 * signature). Does not reject requests by itself - controllers that need
 * auth call AuthUtil.requireUsername, so this stays a plain verify-only filter.
 */
@Component
@Order(1)
public class JwtAuthFilter extends OncePerRequestFilter {

    static final String USERNAME_ATTR = "verifiedUsername";
    static final String ROLE_ATTR = "verifiedRole";

    private final SecretKey key;

    public JwtAuthFilter(@Value("${JWT_SECRET:}") String secret) {
        if (secret == null || secret.isBlank()) {
            throw new IllegalStateException("JWT_SECRET is not set; blog cannot verify tokens");
        }
        this.key = new SecretKeySpec(secret.getBytes(StandardCharsets.UTF_8), "HmacSHA256");
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
            throws ServletException, IOException {
        String header = request.getHeader("Authorization");
        if (header != null && header.regionMatches(true, 0, "Bearer ", 0, 7)) {
            String token = header.substring(7).trim();
            try {
                Claims claims = Jwts.parser().verifyWith(key).build().parseSignedClaims(token).getPayload();
                request.setAttribute(USERNAME_ATTR, claims.getSubject());
                request.setAttribute(ROLE_ATTR, claims.get("role", String.class));
            } catch (JwtException | IllegalArgumentException ex) {
                // invalid/expired token: leave attributes unset, AuthUtil.requireUsername
                // will reject with 401 for any endpoint that needs identity
            }
        }
        filterChain.doFilter(request, response);
    }
}
