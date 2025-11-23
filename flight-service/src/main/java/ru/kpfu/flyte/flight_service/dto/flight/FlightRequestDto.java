package ru.kpfu.flyte.flight_service.dto.flight;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Getter;
import lombok.Setter;
import ru.kpfu.flyte.flight_service.model.FlightStatus;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Getter
@Setter
public class FlightRequestDto {

    @NotBlank
    @Size(max = 20)
    private String flightNumber;

    @NotBlank
    @Size(max = 10)
    private String originAirportCode;

    @NotBlank
    @Size(max = 10)
    private String destinationAirportCode;

    @NotNull
    private LocalDateTime departureTime;

    @NotNull
    private LocalDateTime arrivalTime;

    @NotNull
    private BigDecimal basePrice;

    @NotBlank
    @Size(max = 3)
    private String currency;

    @NotBlank
    @Size(max = 50)
    private String aircraftCode;

    private FlightStatus status;
}
