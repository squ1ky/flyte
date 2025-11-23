package ru.kpfu.flyte.flight_service.mapper;

import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import ru.kpfu.flyte.flight_service.dto.flight.FlightResponseDto;
import ru.kpfu.flyte.flight_service.model.Flight;

import java.util.List;

@Mapper(componentModel = "spring")
public interface FlightMapper {

    @Mapping(target = "originAirportCode", source = "origin.code")
    @Mapping(target = "destinationAirportCode", source = "destination.code")
    @Mapping(target = "aircraftCode", source = "aircraft.code")
    FlightResponseDto toDto(Flight entity);

    List<FlightResponseDto> toDtoList(List<Flight> entities);
}
