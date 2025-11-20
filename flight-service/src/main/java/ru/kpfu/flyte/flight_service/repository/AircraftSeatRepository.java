package ru.kpfu.flyte.flight_service.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.kpfu.flyte.flight_service.model.AircraftSeat;

import java.util.List;

@Repository
public interface AircraftSeatRepository extends JpaRepository<AircraftSeat, Long> {
    List<AircraftSeat> findByAircraftIdOrderByRowNumberAscSeatColumnAsc(Long aircraftId);
}
