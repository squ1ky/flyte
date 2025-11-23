package ru.kpfu.flyte.flight_service.mapper;

import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import ru.kpfu.flyte.flight_service.dto.seat.SeatDto;
import ru.kpfu.flyte.flight_service.model.AircraftSeat;

import java.util.List;

@Mapper(componentModel = "spring")
public interface SeatMapper {

    @Mapping(target = "window", source = "window")
    @Mapping(target = "aisle", source = "aisle")
    @Mapping(target = "exitRow", source = "exitRow")
    SeatDto toDto(AircraftSeat seat);

    List<SeatDto> toDtoList(List<AircraftSeat> seats);
}
