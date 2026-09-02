package com.example.blog.exception;

/** No verified caller identity was present on the request. */
public class UnauthorizedException extends RuntimeException {
    public UnauthorizedException(String message) {
        super(message);
    }
}
