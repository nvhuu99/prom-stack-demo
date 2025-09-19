package com.example.demo_observability.controllers;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class Controller {

    private static final Logger log = LoggerFactory.getLogger(Controller.class);

    @GetMapping
    public ResponseEntity<?> handle() {
        if (Math.random() < 0.2) {
            log.error("Simulate bad request - Path \"/\"");
            return ResponseEntity.badRequest().body("Simulate bad request");
        } else {
            log.info("Request handled - Path: \"/\" - ");
            return ResponseEntity.ok("OK");
        }
    }
}
