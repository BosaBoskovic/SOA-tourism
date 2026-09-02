package com.example.blog.config;

import com.example.blog.exception.UnauthorizedException;
import jakarta.servlet.http.HttpServletRequest;

/** Reads the identity JwtAuthFilter verified for this request. */
public final class AuthUtil {

    private AuthUtil() {
    }

    public static String requireUsername(HttpServletRequest request) {
        Object username = request.getAttribute(JwtAuthFilter.USERNAME_ATTR);
        if (username == null) {
            throw new UnauthorizedException("Authentication required");
        }
        return (String) username;
    }

    public static String usernameOrNull(HttpServletRequest request) {
        Object username = request.getAttribute(JwtAuthFilter.USERNAME_ATTR);
        return username != null ? (String) username : null;
    }
}
