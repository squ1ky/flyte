package ru.kpfu.flyte.flight_service.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.kpfu.flyte.flight_service.model.Airport;

import java.util.List;
import java.util.Optional;

@Repository
public interface AirportRepository extends JpaRepository<Airport, Long> {
    Optional<Airport> findByCode(String code);
    List<Airport> findByCityIgnoreCase(String city);
    List<Airport> findByCityContainingIgnoreCase(String city);
}
