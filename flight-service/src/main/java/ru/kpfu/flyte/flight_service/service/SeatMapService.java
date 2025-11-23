package ru.kpfu.flyte.flight_service.service;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.kpfu.flyte.flight_service.dto.seat.SeatDto;
import ru.kpfu.flyte.flight_service.exception.NotFoundException;
import ru.kpfu.flyte.flight_service.mapper.SeatMapper;
import ru.kpfu.flyte.flight_service.model.Aircraft;
import ru.kpfu.flyte.flight_service.model.AircraftSeat;
import ru.kpfu.flyte.flight_service.model.Flight;
import ru.kpfu.flyte.flight_service.repository.AircraftRepository;
import ru.kpfu.flyte.flight_service.repository.AircraftSeatRepository;
import ru.kpfu.flyte.flight_service.repository.FlightRepository;

import java.util.List;

@Service
@Transactional(readOnly = true)
@RequiredArgsConstructor
public class SeatMapService {

    private final AircraftRepository aircraftRepository;
    private final AircraftSeatRepository aircraftSeatRepository;
    private final FlightRepository flightRepository;
    private final SeatMapper seatMapper;

    public List<SeatDto> getSeatMapByAircraftId(Long aircraftId) {
        Aircraft aircraft = aircraftRepository.findById(aircraftId)
                .orElseThrow(() -> new NotFoundException("Aircraft not found: id=" + aircraftId));

        List<AircraftSeat> seats = aircraftSeatRepository.findByAircraftIdOrderByRowNumberAscSeatColumnAsc(aircraft.getId());
        return seatMapper.toDtoList(seats);
    }

    public List<SeatDto> getSeatMapByFlightId(Long flightId) {
        Flight flight = flightRepository.findById(flightId)
                .orElseThrow(() -> new NotFoundException("Flight not found: id=" + flightId));

        Long aircraftId = flight.getAircraft().getId();
        List<AircraftSeat> seats = aircraftSeatRepository.findByAircraftIdOrderByRowNumberAscSeatColumnAsc(aircraftId);
        return seatMapper.toDtoList(seats);
    }
}
