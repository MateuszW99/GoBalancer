using Microsoft.AspNetCore.Mvc;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddHttpLogging(_ => {});
var app = builder.Build();

app.UseHttpLogging();

app.MapGet("/healthcheck",
    (HttpContext context) => new HealthCheckResponse("1.0.0", DateTime.UtcNow, context.Request.Host.Port));

app.MapPost("/test-endpoint", ([FromBody] TestEndpointRequest request, HttpContext context) =>
        Results.Json(new TestEndpointResponse(context.Request.Host.Port), statusCode: request.RequestedResponseCode));

var ports = Environment.GetEnvironmentVariable("APP_PORTS")?.Split(';') ?? [];
foreach (var port in ports)
    app.Urls.Add($"http://+:{port}");

app.Run();

sealed record HealthCheckResponse(string Version, DateTime CurrentDate, int? Port) : ApiBaseResponse(Port);

sealed record TestEndpointRequest(int RequestedResponseCode);

sealed record TestEndpointResponse(int? Port) : ApiBaseResponse(Port);

record ApiBaseResponse(int? Port);
