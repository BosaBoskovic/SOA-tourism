package com.example.blog.config;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

/**
 * Emits one structured access-log line per HTTP request (method, path,
 * status, duration). Registered after Spring Boot's tracing filter in the
 * chain, so traceId/spanId are already present in the SLF4J MDC (see
 * logback-spring.xml) and end up on the same log line.
 */
@Component
@Order(Integer.MAX_VALUE)
public class AccessLogFilter extends OncePerRequestFilter {

    private static final Logger log = LoggerFactory.getLogger("http.access");

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
            throws ServletException, IOException {
        long start = System.currentTimeMillis();
        try {
            filterChain.doFilter(request, response);
        } finally {
            long durationMs = System.currentTimeMillis() - start;
            log.info("http_request method={} path={} status={} duration_ms={} remote_addr={}",
                    request.getMethod(), request.getRequestURI(), response.getStatus(), durationMs, request.getRemoteAddr());
        }
    }
}
