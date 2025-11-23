package ru.kpfu.flyte.flight_service.dto.flight;

import jakarta.validation.constraints.NotNull;
import lombok.Getter;
import lombok.Setter;
import ru.kpfu.flyte.flight_service.model.FlightStatus;

@Getter
@Setter
public class FlightStatusUpdateDto {

    @NotNull
    private FlightStatus status;
}
