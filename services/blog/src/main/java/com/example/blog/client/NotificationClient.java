package com.example.blog.client;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;
import org.springframework.web.client.RestClientException;

import java.util.Map;

/**
 * Best-effort call to stakeholders' internal notification endpoint. Never
 * allowed to break the comment/like flow that triggers it.
 */
@Component
public class NotificationClient {

    private static final Logger log = LoggerFactory.getLogger(NotificationClient.class);

    private final RestClient restClient;

    public NotificationClient(@Value("${stakeholders.service.url}") String stakeholdersUrl) {
        SimpleClientHttpRequestFactory requestFactory = new SimpleClientHttpRequestFactory();
        requestFactory.setConnectTimeout(2000);
        requestFactory.setReadTimeout(3000);
        this.restClient = RestClient.builder()
                .baseUrl(stakeholdersUrl)
                .requestFactory(requestFactory)
                .build();
    }

    public void notifyNewComment(String blogAuthorUsername, String commenterUsername) {
        if (blogAuthorUsername.equals(commenterUsername)) {
            return; // no need to notify yourself
        }
        send(blogAuthorUsername, "comment", commenterUsername + " je komentarisao/la tvoj blog.", commenterUsername);
    }

    private void send(String username, String type, String message, String relatedUsername) {
        try {
            restClient.post()
                    .uri("/stakeholders/notifications/internal")
                    .body(Map.of(
                            "username", username,
                            "type", type,
                            "message", message,
                            "relatedUsername", relatedUsername
                    ))
                    .retrieve()
                    .toBodilessEntity();
        } catch (RestClientException ex) {
            log.warn("notification delivery failed for {}: {}", username, ex.getMessage());
        }
    }
}
