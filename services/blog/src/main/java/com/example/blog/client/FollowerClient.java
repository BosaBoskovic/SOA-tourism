package com.example.blog.client;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;
import org.springframework.web.client.RestClientException;

import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

@Component
public class FollowerClient {

    private static final Logger log = LoggerFactory.getLogger(FollowerClient.class);

    private final RestClient restClient;

    public FollowerClient(@Value("${follower.service.url}") String followerServiceUrl) {
        SimpleClientHttpRequestFactory requestFactory = new SimpleClientHttpRequestFactory();
        requestFactory.setConnectTimeout(2000);
        requestFactory.setReadTimeout(3000);
        this.restClient = RestClient.builder()
                .baseUrl(followerServiceUrl)
                .requestFactory(requestFactory)
                .build();
    }

    // Fails closed: if followers is unreachable/slow, treat as "not following"
    // rather than letting an exception bubble up and break the whole request.
    public boolean isFollowing(String followerUsername, String targetUsername) {
        if (followerUsername == null || targetUsername == null) {
            return false;
        }

        if (followerUsername.equals(targetUsername)) {
            return true;
        }

        try {
            Map<String, Object> response = restClient.get()
                    .uri(uriBuilder -> uriBuilder
                            .path("/followers/is-following")
                            .queryParam("followerUsername", followerUsername)
                            .queryParam("targetUsername", targetUsername)
                            .build())
                    .retrieve()
                    .body(Map.class);

            if (response == null || !response.containsKey("isFollowing")) {
                return false;
            }

            Object value = response.get("isFollowing");
            if (value instanceof Boolean follows) {
                return follows;
            }
            return Boolean.parseBoolean(String.valueOf(value));
        } catch (RestClientException ex) {
            log.warn("followers service unreachable, treating {} as not following {}: {}", followerUsername, targetUsername, ex.getMessage());
            return false;
        }
    }

    // Falls back to "only your own posts" if followers is unreachable/slow,
    // instead of breaking the whole blog list.
    public Set<String> getVisibleAuthors(String username) {
        if (username == null || username.isBlank()) {
            return Set.of();
        }

        try {
            Map<String, Object> response = restClient.get()
                    .uri("/followers/visible-authors/{username}", username)
                    .retrieve()
                    .body(Map.class);

            if (response == null) {
                return Set.of(username);
            }

            Object rawAuthors = response.get("authors");
            if (!(rawAuthors instanceof List<?> authorsList)) {
                return Set.of(username);
            }

            Set<String> authors = new HashSet<>();
            for (Object item : authorsList) {
                if (item != null) {
                    authors.add(String.valueOf(item));
                }
            }

            authors.add(username);
            return authors;
        } catch (RestClientException ex) {
            log.warn("followers service unreachable, falling back to {}'s own posts only: {}", username, ex.getMessage());
            return Set.of(username);
        }
    }
}
