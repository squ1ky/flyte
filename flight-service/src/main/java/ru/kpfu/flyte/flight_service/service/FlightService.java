package ru.kpfu.flyte.flight_service.service;

import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.kpfu.flyte.flight_service.dto.flight.FlightRequestDto;
import ru.kpfu.flyte.flight_service.dto.flight.FlightResponseDto;
import ru.kpfu.flyte.flight_service.dto.flight.FlightSearchCriteria;
import ru.kpfu.flyte.flight_service.dto.flight.FlightStatusUpdateDto;
import ru.kpfu.flyte.flight_service.exception.NotFoundException;
import ru.kpfu.flyte.flight_service.exception.ValidationException;
import ru.kpfu.flyte.flight_service.mapper.FlightMapper;
import ru.kpfu.flyte.flight_service.model.Aircraft;
import ru.kpfu.flyte.flight_service.model.Airport;
import ru.kpfu.flyte.flight_service.model.Flight;
import ru.kpfu.flyte.flight_service.model.FlightStatus;
import ru.kpfu.flyte.flight_service.repository.AircraftRepository;
import ru.kpfu.flyte.flight_service.repository.AirportRepository;
import ru.kpfu.flyte.flight_service.repository.FlightRepository;

import java.time.LocalDate;
import java.time.LocalDateTime;

@Service
@Transactional
@RequiredArgsConstructor
public class FlightService {

    private final FlightRepository flightRepository;
    private final AirportRepository airportRepository;
    private final AircraftRepository aircraftRepository;
    private final FlightMapper flightMapper;

    public FlightResponseDto createFlight(FlightRequestDto request) {
        if (request.getArrivalTime().isBefore(request.getDepartureTime()) ||
            request.getArrivalTime().isEqual(request.getDepartureTime())) {
            throw new ValidationException("arrivalTime must be after departureTime");
        }

        Airport origin = airportRepository.findByCode(request.getOriginAirportCode())
                .orElseThrow(() -> new NotFoundException("Origin airport not found: code=" + request.getOriginAirportCode()));

        Airport destination = airportRepository.findByCode(request.getDestinationAirportCode())
                .orElseThrow(() -> new NotFoundException("Destination airport not found: code=" + request.getDestinationAirportCode()));

        Aircraft aircraft = aircraftRepository.findByCode(request.getAircraftCode())
                .orElseThrow(() -> new NotFoundException("Aircraft not found: code=" + request.getAircraftCode()));

        Flight flight = Flight.builder()
                .flightNumber(request.getFlightNumber())
                .origin(origin)
                .destination(destination)
                .departureTime(request.getDepartureTime())
                .arrivalTime(request.getArrivalTime())
                .basePrice(request.getBasePrice())
                .currency(request.getCurrency())
                .aircraft(aircraft)
                .status(request.getStatus() != null ? request.getStatus() : FlightStatus.SCHEDULED)
                .build();

        Flight saved = flightRepository.save(flight);
        return flightMapper.toDto(saved);
    }

    @Transactional(readOnly = true)
    public FlightResponseDto getFlight(Long id) {
        Flight flight = flightRepository.findById(id)
                .orElseThrow(() -> new NotFoundException("Flight not found: id=" + id));
        return flightMapper.toDto(flight);
    }

    @Transactional(readOnly = true)
    public Page<FlightResponseDto> searchFlights(FlightSearchCriteria criteria, Pageable pageable) {
        if (criteria.getOriginAirportCode() == null || criteria.getDestinationAirportCode() == null ||
            criteria.getDepartureDate() == null) {
            throw new ValidationException("originAirportCode, destinationAirportCode and departureDate are required");
        }

        Airport origin = airportRepository.findByCode(criteria.getOriginAirportCode())
                .orElseThrow(() -> new NotFoundException("Origin airport not found: code=" + criteria.getOriginAirportCode()));

        Airport destination = airportRepository.findByCode(criteria.getDestinationAirportCode())
                .orElseThrow(() -> new NotFoundException("Destination airport not found: code=" + criteria.getDestinationAirportCode()));

        LocalDate date = criteria.getDepartureDate().toLocalDate();
        LocalDateTime from = date.atStartOfDay();
        LocalDateTime to = date.plusDays(1).atStartOfDay();

        Page<Flight> page = flightRepository.findByOriginIdAndDestinationIdAndDepartureTimeBetween(
                origin.getId(),
                destination.getId(),
                from,
                to,
                pageable
        );

        return page.map(flightMapper::toDto);
    }

    public FlightResponseDto updateStatus(Long flightId, FlightStatusUpdateDto dto) {
        Flight flight = flightRepository.findById(flightId)
                .orElseThrow(() -> new NotFoundException("Flight not found: id=" + flightId));

        flight.setStatus(dto.getStatus());
        Flight saved = flightRepository.save(flight);
        return flightMapper.toDto(saved);
    }
}
