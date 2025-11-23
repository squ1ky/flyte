package ru.kpfu.flyte.flight_service.mapper;

import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.MappingTarget;
import ru.kpfu.flyte.flight_service.dto.airport.AirportRequestDto;
import ru.kpfu.flyte.flight_service.dto.airport.AirportResponseDto;
import ru.kpfu.flyte.flight_service.model.Airport;

import java.util.List;

@Mapper(componentModel = "spring")
public interface AirportMapper {

    AirportResponseDto toDto(Airport entity);

    List<AirportResponseDto> toDtoList(List<Airport> entities);

    @Mapping(target = "id", ignore = true)
    @Mapping(target = "createdAt", ignore = true)
    @Mapping(target = "updatedAt", ignore = true)
    Airport toEntity(AirportRequestDto dto);

    @Mapping(target = "id", ignore = true)
    @Mapping(target = "code", ignore = true)
    @Mapping(target = "createdAt", ignore = true)
    @Mapping(target = "updatedAt", ignore = true)
    void updateAirportFromDto(AirportRequestDto dto, @MappingTarget Airport entity);
}
