/*
Package server provides an NTRIP server implementation for streaming RTCM v3.3 data.

The NTRIP (Networked Transport of RTCM via Internet Protocol) server is responsible for
connecting to an NTRIP caster and publishing RTCM correction data. This package implements
a complete NTRIP server following the NTRIP protocol specification.

# Main Components

## Server

The Server type is the main component of this package. It connects to an NTRIP caster
and publishes RTCM data from a data source.

Example usage:

    // Create a new server
    server := server.NewServer("example.com", "2101", "username", "password", "MOUNTPOINT")
    
    // Set the data source
    server.SetDataSource(dataSource)
    
    // Start the server
    if err := server.Start(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
    defer server.Stop()

## DataSource

The DataSource interface defines the methods that a data source must implement.
A data source is responsible for providing RTCM data to the server.

# SOLID Principles

This package follows the SOLID principles:

- Single Responsibility: Each component has a single responsibility
- Open/Closed: Components are open for extension but closed for modification
- Liskov Substitution: Interfaces can be substituted with their implementations
- Interface Segregation: Interfaces are specific to client needs
- Dependency Inversion: High-level modules depend on abstractions

# Thread Safety

The server is designed to be thread-safe and can handle multiple concurrent connections.

# Protocol Support

The server supports both NTRIP version 1 and 2 protocols.
*/
package server
