package ru.kpfu.flyte.flight_service.service;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import ru.kpfu.flyte.flight_service.dto.aircraft.AircraftRequestDto;
import ru.kpfu.flyte.flight_service.dto.aircraft.AircraftResponseDto;
import ru.kpfu.flyte.flight_service.exception.NotFoundException;
import ru.kpfu.flyte.flight_service.exception.ValidationException;
import ru.kpfu.flyte.flight_service.mapper.AircraftMapper;
import ru.kpfu.flyte.flight_service.model.Aircraft;
import ru.kpfu.flyte.flight_service.repository.AircraftRepository;

import java.util.List;

@Service
@Transactional
@RequiredArgsConstructor
public class AircraftService {

    private final AircraftRepository aircraftRepository;
    private final AircraftMapper aircraftMapper;

    public AircraftResponseDto createAircraft(AircraftRequestDto request) {
        aircraftRepository.findByCode(request.getCode())
                .ifPresent(a -> {
                    throw new ValidationException("Aircraft with code '" + request.getCode() + "' already exists");
                });

        Aircraft aircraft = aircraftMapper.toEntity(request);
        Aircraft saved = aircraftRepository.save(aircraft);
        return aircraftMapper.toDto(saved);
    }

    @Transactional(readOnly = true)
    public AircraftResponseDto getAircraft(Long id) {
        Aircraft aircraft = aircraftRepository.findById(id)
                .orElseThrow(() -> new NotFoundException("Aircraft not found: id=" + id));
        return aircraftMapper.toDto(aircraft);
    }

    @Transactional(readOnly = true)
    public List<AircraftResponseDto> getAllAircrafts() {
        return aircraftMapper.toDtoList(aircraftRepository.findAll());
    }
}
