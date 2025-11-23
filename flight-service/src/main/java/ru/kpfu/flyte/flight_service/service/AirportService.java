package ru.kpfu.flyte.flight_service.service;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.kpfu.flyte.flight_service.dto.airport.AirportRequestDto;
import ru.kpfu.flyte.flight_service.dto.airport.AirportResponseDto;
import ru.kpfu.flyte.flight_service.exception.NotFoundException;
import ru.kpfu.flyte.flight_service.mapper.AirportMapper;
import ru.kpfu.flyte.flight_service.model.Airport;
import ru.kpfu.flyte.flight_service.repository.AirportRepository;

import java.util.List;

@Service
@Transactional
@RequiredArgsConstructor
public class AirportService {

    private final AirportRepository airportRepository;
    private final AirportMapper airportMapper;

    public AirportResponseDto createAirport(AirportRequestDto request) {
        airportRepository.findByCode(request.getCode())
                .ifPresent(a -> {
                    throw new IllegalArgumentException("Airport with code '" + request.getCode() + "' already exists");
                });

        Airport airport = airportMapper.toEntity(request);
        Airport saved = airportRepository.save(airport);
        return airportMapper.toDto(saved);
    }

    @Transactional(readOnly = true)
    public AirportResponseDto getAirport(Long id) {
        Airport airport = airportRepository.findById(id)
                .orElseThrow(() -> new NotFoundException("Airport not found: id=" + id));
        return airportMapper.toDto(airport);
    }

    @Transactional(readOnly = true)
    public List<AirportResponseDto> findAirports(String code, String city) {
        List<Airport> airports;
        if (code != null && !code.isBlank()) {
            airports = airportRepository.findByCode(code)
                    .map(List::of)
                    .orElseGet(List::of);
        } else if (city != null && !city.isBlank()) {
            airports = airportRepository.findByCityContainingIgnoreCase(city);
        } else {
            airports = airportRepository.findAll();
        }

        return airportMapper.toDtoList(airports);
    }
}
