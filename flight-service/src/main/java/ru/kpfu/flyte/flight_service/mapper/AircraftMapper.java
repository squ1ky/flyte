package ru.kpfu.flyte.flight_service.mapper;

import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import ru.kpfu.flyte.flight_service.dto.aircraft.AircraftRequestDto;
import ru.kpfu.flyte.flight_service.dto.aircraft.AircraftResponseDto;
import ru.kpfu.flyte.flight_service.model.Aircraft;

import java.util.List;

@Mapper(componentModel = "spring")
public interface AircraftMapper {

    AircraftResponseDto toDto(Aircraft entity);

    List<AircraftResponseDto> toDtoList(List<Aircraft> entities);

    @Mapping(target = "id", ignore = true)
    @Mapping(target = "createdAt", ignore = true)
    @Mapping(target = "updatedAt", ignore = true)
    @Mapping(target = "seats", ignore = true)
    Aircraft toEntity(AircraftRequestDto dto);
}
