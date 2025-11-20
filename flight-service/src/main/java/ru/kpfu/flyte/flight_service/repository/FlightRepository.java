package ru.kpfu.flyte.flight_service.repository;

import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.kpfu.flyte.flight_service.model.Flight;
import ru.kpfu.flyte.flight_service.model.FlightStatus;

import java.time.LocalDateTime;

@Repository
public interface FlightRepository extends JpaRepository<Flight, Long> {
    Page<Flight> findByFlightNumber(String flightNumber, Pageable pageable);
    Page<Flight> findByOriginIdAndDestinationIdAndDepartureTimeBetween(
            Long originId,
            Long destinationId,
            LocalDateTime departureFrom,
            LocalDateTime departureTo,
            Pageable pageable
    );
    Page<Flight> findByOriginIdAndDestinationIdAndDepartureTimeBetweenAndStatus(
            Long originId,
            Long destinationId,
            LocalDateTime departureFrom,
            LocalDateTime departureTo,
            FlightStatus status,
            Pageable pageable
    );
}
