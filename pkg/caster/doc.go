/*
Package caster provides an NTRIP caster implementation for streaming RTCM v3.3 data.

The NTRIP (Networked Transport of RTCM via Internet Protocol) caster acts as a central hub
that receives RTCM correction data from one or more NTRIP servers and distributes it to
multiple NTRIP clients. This package implements a complete NTRIP caster following the
NTRIP protocol specification.

# Main Components

## Caster

The Caster type is the main component of this package. It provides an HTTP server that
handles NTRIP client and server connections, manages the sourcetable, and routes data
between servers and clients.

Example usage:

    // Create a new source service
    svc := inmemory.NewInMemorySourceService()
    
    // Add mountpoints to the sourcetable
    svc.Sourcetable.Mounts = append(svc.Sourcetable.Mounts, caster.StreamEntry{
        Name:       "EXAMPLE",
        Identifier: "EXAMPLE",
        Format:     "RTCM 3.3",
    })
    
    // Create a new caster
    caster := caster.NewCaster(":2101", svc, logger)
    
    // Start the caster
    if err := caster.ListenAndServe(); err != nil {
        log.Fatalf("Caster error: %v", err)
    }

## SourceService

The SourceService interface defines the methods that a source service must implement.
A source service is responsible for managing the sourcetable and handling publisher and
subscriber connections.

## Sourcetable

The Sourcetable type represents the NTRIP sourcetable, which contains information about
the available mountpoints, networks, and the caster itself.

# SOLID Principles

This package follows the SOLID principles:

- Single Responsibility: Each component has a single responsibility
- Open/Closed: Components are open for extension but closed for modification
- Liskov Substitution: Interfaces can be substituted with their implementations
- Interface Segregation: Interfaces are specific to client needs
- Dependency Inversion: High-level modules depend on abstractions

# Thread Safety

The caster is designed to be thread-safe and can handle multiple concurrent connections.

# Protocol Support

The caster supports both NTRIP version 1 and 2 protocols.
*/
package caster
